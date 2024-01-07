package main

import (
	//"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	common "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	events "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/msgs2"
	"github.com/remeh/sizedwaitgroup"
)

//TODO
//"Catch up on the score" - dont remember what this is lol

//BUG fix
//MAKE ROUNDENDOFFICIAL Redundant (may have done, make sure it passes validation tho)
//add verification for missed event triggers if someone DCs/Cs after the redundant event that is stale
//csgo bots all have same steamID, need to use something else just in case for bots
//MM bug?

//add support for esea games? need to change validation and how we determine what round it is (gamestats rounds doesnt work)

//FUNCTIONAL CHANGES
//add verification for if a round event has triggered so far in the round (avoid double roundEnds)
//check for game start without pistol (if we have bad demo)
//Add backend support
//Add anchor stuff
//Add team economy round stats (ecos, forces, etc)
//Add various nil checking

//CLEAN CODE
//TODO: create a outputPlayer function to clean up output.go
//TODO: convert rating calculations to a function
//TODO: actually use killValues lmao

const DEBUG = true

//const suppressNormalOutput = false

// globals
const printChatLog = true
const printDebugLog = true
const FORCE_NEW_STATS_UPLOAD = false
const ENABLE_WPA_DATA_OUTPUT = false
const BACKEND_PUSHING = false

const tradeCutoff = 4 // in seconds
var multikillBonus = [...]float64{0, 0, 0.3, 0.7, 1.2, 2}
var clutchBonus = [...]float64{0, 0.2, 0.6, 1.2, 2, 3}
var killValues = map[string]float64{
	"attacking":     1.2, //base values
	"defending":     1.0,
	"bombDefense":   1.0,
	"retake":        1.2,
	"chase":         0.8,
	"exit":          0.6,
	"t_consolation": 0.5,
	"gravy":         0.6,
	"punish":        0.8,
	"entry":         0.8, //multipliers
	"t_opener":      0.3,
	"ct_opener":     0.5,
	"trade":         0.3,
	"flashAssist":   0.2,
	"assist":        0.15,
}

func main() {
	swg := sizedwaitgroup.New(2)
	input_dir := "in"
	files, _ := os.ReadDir(input_dir)

	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, ".dem") {
			swg.Add()
			fmt.Println("processing", file.Name())
			go processDemo(filepath.Join(input_dir, filename), &swg)
		}
	}

	swg.Wait()
	os.Exit(0)
}

func initGameObject() *game {
	g := game{
		reconnectedPlayers: make(map[uint64]bool),
	}
	g.rounds = make([]*round, 0)
	g.potentialRound = &round{}

	g.flags.hasGameStarted = false
	g.flags.isGameLive = false
	g.flags.isGameOver = false
	g.flags.inRound = false
	g.flags.prePlant = true
	g.flags.postPlant = false
	g.flags.postWinCon = false
	//these three vars to check if we have a complete round
	g.flags.roundIntegrityStart = -1
	g.flags.roundIntegrityEnd = -1
	g.flags.roundIntegrityEndOfficial = -1

	return &g
}

func processDemo(demoName string, swg *sizedwaitgroup.SizedWaitGroup) {
	//defer swg.Done()

	game := initGameObject()

	/*
	   var errLog = "";
	*/

	f, err := os.Open(demoName)
	//f, err := os.Open("league1.dem")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	tempCoreID := strings.Split(demoName, "mid")
	if len(tempCoreID) > 1 {
		//we have a CSC match
		parsed := strings.Split(strings.Split(strings.Split(demoName, "mid")[1], ".")[0], "-")
		game.coreID = parsed[0]
		//game.mapNum, _ = strconv.Atoi(strings.Split(parsed[1], "_")[0])
		game.mapNum = 0 //this will break the backend, is dep'd
	}

	p := dem.NewParser(f)
	defer p.Close()

	//must parse header to get header info
	header, err := p.ParseHeader()
	if err != nil {
		panic(err)
	}

	//set tick rate
	game.tickRate = int(math.Round(p.TickRate()))
	fmt.Println("Tick rate is", game.tickRate, "| Map is", game.mapName)

	game.tickLength = header.PlaybackTicks

	//creating a file to dump chat log into
	os.Mkdir("out", 0777)
	chatFile, chatFileErr := os.Create("out/chatLog.txt")
	if chatFileErr != nil {
		fmt.Printf("Error in opening chatLog file.\n")
	}
	defer chatFile.Close()

	debugFile, debugFileErr := os.Create("out/debug.txt")
	if debugFileErr != nil {
		fmt.Printf("Error in opening debug file.\n")
	}
	defer debugFile.Close()

	//---------------FUNCTIONS---------------

	initGameStart := func() {
		game.flags.hasGameStarted = true
		game.flags.isGameLive = true
		fmt.Println("GAME HAS STARTED!!!")

		// In case the tickRate is 0 we want to re-set it based on the tickInterval now that the game has hasGameStarted
		if game.tickRate == 0 {
			game.tickRate = int(math.Round(p.TickRate()))
		}

		game.teams = make(map[string]*team)

		teamTemp := p.GameState().TeamTerrorists()
		game.teams[validateTeamName(game, teamTemp.ClanName(), teamTemp.Team())] = &team{name: validateTeamName(game, teamTemp.ClanName(), teamTemp.Team())}
		teamTemp = p.GameState().TeamCounterTerrorists()
		game.teams[validateTeamName(game, teamTemp.ClanName(), teamTemp.Team())] = &team{name: validateTeamName(game, teamTemp.ClanName(), teamTemp.Team())}

		//to handle short and long matches
		//TODO: DECIDE IF THIS IS STILL NEEDED
		//if p.GameState().ConVars()["mp_maxrounds"] != "30" {
		//	maxRounds, fuckOFF := strconv.Atoi(p.GameState().ConVars()["mp_maxrounds"])
		//	if fuckOFF == nil {
		//		game.roundsToWin = maxRounds/2 + 1
		//	} else {
		//		//ADD TO ERROR LOG
		//		game.roundsToWin = maxRounds/2 + 1
		//		//maybe this gives us a way to check for short vs long match
		//	}
		//} else {
		//	game.roundsToWin = 16 //we will assume long match in case convar is not set
		//}
		game.roundsToWin = 13

	}

	//reset various flags
	resetRoundFlags := func() {
		game.flags.prePlant = true
		game.flags.postPlant = false
		game.flags.postWinCon = false
		game.flags.tClutchVal = 0
		game.flags.ctClutchVal = 0
		game.flags.tClutchSteam = 0
		game.flags.ctClutchSteam = 0
		game.flags.tMoney = false
		game.flags.openingKill = true
		game.flags.lastTickProcessed = 0
		game.flags.ticksProcessed = 0
		game.flags.didRoundEndFire = false
		game.flags.roundStartedAt = 0
		game.flags.haveInitRound = false
	}

	initTeamPlayer := func(team *common.TeamState, currRoundObj *round) {
		for _, teamMember := range getTeamMembers(team, game, p) {
			player := &playerStats{name: teamMember.Name, steamID: teamMember.SteamID64, isBot: teamMember.IsBot, side: int(team.Team()), teamENUM: team.ID(), teamClanName: validateTeamName(game, team.ClanName(), team.Team()), health: 100, tradeList: make(map[uint64]int), damageList: make(map[uint64]int)}
			currRoundObj.playerStats[player.steamID] = player
		}
	}

	initRound := func() {
		game.flags.roundIntegrityStart = p.GameState().TotalRoundsPlayed() + 1
		if DEBUG {
			fmt.Println("We are starting round", game.flags.roundIntegrityStart)
		}

		newRound := &round{roundNum: int8(game.flags.roundIntegrityStart), startingTick: p.GameState().IngameTick()}
		newRound.playerStats = make(map[uint64]*playerStats)
		newRound.teamStats = make(map[string]*teamStats)

		//set players in playerStats for the round
		terrorists := p.GameState().TeamTerrorists()
		counterTerrorists := p.GameState().TeamCounterTerrorists()

		initTeamPlayer(terrorists, newRound)
		initTeamPlayer(counterTerrorists, newRound)

		//set teams in teamStats for the round
		newRound.teamStats[validateTeamName(game, p.GameState().TeamTerrorists().ClanName(), p.GameState().TeamTerrorists().Team())] = &teamStats{tR: 1}
		newRound.teamStats[validateTeamName(game, p.GameState().TeamCounterTerrorists().ClanName(), p.GameState().TeamCounterTerrorists().Team())] = &teamStats{ctR: 1}

		//create empty WPAlog
		newRound.WPAlog = make([]*wpalog, 0)

		// Reset round
		game.potentialRound = newRound

		//track the number of people alive for clutch checking and record keeping
		game.flags.tAlive = len(getTeamMembers(terrorists, game, p))
		game.flags.ctAlive = len(getTeamMembers(counterTerrorists, game, p))
		game.potentialRound.initTerroristCount = game.flags.tAlive
		game.potentialRound.initCTerroristCount = game.flags.ctAlive

		resetRoundFlags()
	}

	processRoundOnWinCon := func(winnerClanName string) {
		//TODO: FIGURE OUT IF THIS IS CORRECT. I REMOVED +1 FROM TOTALROUNDSPLAYED
		game.flags.roundIntegrityEnd = p.GameState().TotalRoundsPlayed()
		if DEBUG {
			fmt.Println("We are processing round win con stuff", game.flags.roundIntegrityEnd)
		}

		game.totalRounds = game.flags.roundIntegrityEnd

		game.flags.prePlant = false
		game.flags.postPlant = false
		game.flags.postWinCon = true

		//set winner
		game.potentialRound.winnerClanName = winnerClanName
		//fmt.Println("We think this team won", winnerClanName)
		if !game.potentialRound.knifeRound {
			game.teams[game.potentialRound.winnerClanName].score += 1
		}
		//go through and set our WPAlog output to the winner
		for _, log := range game.potentialRound.WPAlog {
			log.winner = game.potentialRound.winnerENUM - 2
		}
	}

	processRoundFinal := func(lastRound bool) {
		game.flags.inRound = false
		game.potentialRound.endingTick = p.GameState().IngameTick()
		game.flags.roundIntegrityEndOfficial = p.GameState().TotalRoundsPlayed()
		if lastRound {
			//game.flags.roundIntegrityEndOfficial += 1
			game.totalRounds = game.flags.roundIntegrityEndOfficial
		}
		if DEBUG {
			fmt.Println("We are processing round final stuff", game.flags.roundIntegrityEndOfficial)
			fmt.Println(len(game.rounds))
		}

		//we have the entire round uninterrupted
		if game.flags.roundIntegrityStart == game.flags.roundIntegrityEnd && game.flags.roundIntegrityEnd == game.flags.roundIntegrityEndOfficial {
			game.potentialRound.integrityCheck = true

			//check team stats
			if game.potentialRound.teamStats[game.potentialRound.winnerClanName].pistols == 1 {
				game.potentialRound.teamStats[game.potentialRound.winnerClanName].pistolsW = 1
			}
			if game.potentialRound.teamStats[game.potentialRound.winnerClanName]._4v5s == 1 {
				game.potentialRound.teamStats[game.potentialRound.winnerClanName]._4v5w = 1
			} else if game.potentialRound.teamStats[game.potentialRound.winnerClanName]._5v4s == 1 {
				game.potentialRound.teamStats[game.potentialRound.winnerClanName]._5v4w = 1
			}
			if game.potentialRound.teamStats[game.potentialRound.winnerClanName].tR == 1 {
				game.potentialRound.teamStats[game.potentialRound.winnerClanName].tRW = 1
			} else if game.potentialRound.teamStats[game.potentialRound.winnerClanName].ctR == 1 {
				game.potentialRound.teamStats[game.potentialRound.winnerClanName].ctRW = 1
			}

			//set the clutch
			if game.potentialRound.winnerENUM == 2 && game.flags.tClutchSteam != 0 {
				game.potentialRound.teamStats[game.potentialRound.winnerClanName].clutches = 1
				game.potentialRound.playerStats[game.flags.tClutchSteam].impactPoints += clutchBonus[game.flags.tClutchVal]
				switch game.flags.tClutchVal {
				case 1:
					game.potentialRound.playerStats[game.flags.tClutchSteam].cl_1 = 1
				case 2:
					game.potentialRound.playerStats[game.flags.tClutchSteam].cl_2 = 1
				case 3:
					game.potentialRound.playerStats[game.flags.tClutchSteam].cl_3 = 1
				case 4:
					game.potentialRound.playerStats[game.flags.tClutchSteam].cl_4 = 1
				case 5:
					game.potentialRound.playerStats[game.flags.tClutchSteam].cl_5 = 1
				}
			} else if game.potentialRound.winnerENUM == 3 && game.flags.ctClutchSteam != 0 {
				game.potentialRound.teamStats[game.potentialRound.winnerClanName].clutches = 1
				game.potentialRound.playerStats[game.flags.ctClutchSteam].impactPoints += clutchBonus[game.flags.ctClutchVal]
				switch game.flags.ctClutchVal {
				case 1:
					game.potentialRound.playerStats[game.flags.ctClutchSteam].cl_1 = 1
				case 2:
					game.potentialRound.playerStats[game.flags.ctClutchSteam].cl_2 = 1
				case 3:
					game.potentialRound.playerStats[game.flags.ctClutchSteam].cl_3 = 1
				case 4:
					game.potentialRound.playerStats[game.flags.ctClutchSteam].cl_4 = 1
				case 5:
					game.potentialRound.playerStats[game.flags.ctClutchSteam].cl_5 = 1
				}
			}

			//add multikills & saves & misc
			highestImpactPoints := 0.0
			mipPlayers := 0
			for _, player := range (game.potentialRound).playerStats {
				if player.deaths == 0 {
					player.kastRounds = 1
					if player.teamENUM != game.potentialRound.winnerENUM {
						player.saves = 1
						game.potentialRound.teamStats[player.teamClanName].saves = 1
					}
				}
				game.potentialRound.playerStats[player.steamID].impactPoints += player.killPoints
				game.potentialRound.playerStats[player.steamID].impactPoints += float64(player.damage) / float64(250)
				game.potentialRound.playerStats[player.steamID].impactPoints += multikillBonus[player.kills]

				switch player.kills {
				case 2:
					player._2k = 1
				case 3:
					player._3k = 1
				case 4:
					player._4k = 1
				case 5:
					player._5k = 1
				}

				if player.impactPoints > highestImpactPoints {
					highestImpactPoints = player.impactPoints
				}

				if player.teamENUM == game.potentialRound.winnerENUM {
					player.winPoints = player.impactPoints

					player.RF = 1
				} else {
					player.RA = 1
				}
			}

			for _, player := range (game.potentialRound).playerStats {
				if player.impactPoints == highestImpactPoints {
					mipPlayers += 1
				}
			}
			for _, player := range (game.potentialRound).playerStats {
				if player.impactPoints == highestImpactPoints {
					player.mip = 1.0 / float64(mipPlayers)
				}
			}

			//check the lurk
			var susLurker uint64
			susLurkBlips := 0
			invalidLurk := false
			for _, player := range game.potentialRound.playerStats {
				if player.side == 2 {
					if player.lurkerBlips > susLurkBlips {
						susLurkBlips = player.lurkerBlips
						susLurker = player.steamID
					}
				}
			}
			for _, player := range game.potentialRound.playerStats {
				if player.side == 2 {
					if player.lurkerBlips == susLurkBlips && player.steamID != susLurker {
						invalidLurk = true
					}
				}
			}
			if !invalidLurk && susLurkBlips > 3 {
				game.potentialRound.playerStats[susLurker].lurkRounds = 1
			}

			//add our valid round
			game.rounds = append(game.rounds, game.potentialRound)
			if DEBUG {
				fmt.Println("We are appending a round")
			}

		}

		//endRound function functionality

	}

	//-------------ALL OUR EVENTS---------------------

	p.RegisterNetMessageHandler(func(msg *msgs2.CSVCMsg_ServerInfo) {
		game.mapName = *msg.MapName
	})

	p.RegisterEventHandler(func(e events.PlayerInfo) {
		fmt.Println("PlayerInfo", e)
		player := p.GameState().Participants().AllByUserID()[e.Index]
		fmt.Println("PlayerInfo", p.GameState().Participants().AllByUserID())
		if player != nil {
			fmt.Println("PlayerInfo", player)
			// if game.potentialRound.playerStats[player.SteamID64] == nil {
			// 	game.potentialRound.playerStats[player.SteamID64] = &playerStats{name: player.Name, steamID: player.SteamID64, isBot: player.IsBot, side: int(player.Team), teamENUM: int(player.Team), teamClanName: validateTeamName(game, player.TeamState.ClanName(), player.TeamState.Team()), health: 100, tradeList: make(map[uint64]int), damageList: make(map[uint64]int)}
			// }
			game.reconnectedPlayers[player.SteamID64] = true
		}
	})

	p.RegisterEventHandler(func(e events.FrameDone) {
		//fmt.Println("DIBES ", game.flags.isGameLive)
		if game.flags.roundStartedAt > 0 && game.flags.roundStartedAt+(1*game.tickRate) > p.GameState().IngameTick() && !game.flags.haveInitRound {
			pistol := false

			//we are going to check to see if the first pistol is actually starting
			membersT := getTeamMembers(p.GameState().TeamTerrorists(), game, p)
			membersCT := getTeamMembers(p.GameState().TeamCounterTerrorists(), game, p)
			if len(membersT) != 0 && len(membersCT) != 0 {
				if membersT[0].Money()+membersT[0].MoneySpentThisRound() == 800 && membersCT[0].Money()+membersCT[0].MoneySpentThisRound() == 800 {
					//start the game
					if !game.flags.hasGameStarted {
						initGameStart()
					}

					//track the pistol
					pistol = true
				}
			}
			//fmt.Println("Has the Game Started?", game.flags.hasGameStarted)

			if game.flags.isGameLive {
				//init round stats
				initRound()
				game.flags.haveInitRound = true
				if pistol {
					for _, team := range game.potentialRound.teamStats {
						team.pistols = 1
					}
				}

			}
		}

		//add to WPAlog
		if game.flags.inRound && !game.flags.postWinCon && ENABLE_WPA_DATA_OUTPUT {
			//hits every new frame (typically each 1-4 ticks)
			logSize := len(game.potentialRound.WPAlog)
			clock := 0
			planted := 0
			if game.potentialRound.planter != 0 {
				planted = 1
				bombTime, boo := p.GameState().Rules().BombTime()
				bombClock := 0
				if boo != nil {
					bombClock = 40
				} else {
					bombClock = int(bombTime.Seconds())
				}
				clock = bombClock - ((p.GameState().IngameTick() - game.potentialRound.bombStartTick) / game.tickRate)
			} else {
				roundTime, boo := p.GameState().Rules().RoundTime()
				if boo != nil {
					fmt.Println("RUROO RAGGY")
				}
				clock = int(roundTime.Seconds()) - ((p.GameState().IngameTick() - game.potentialRound.startingTick) / game.tickRate)
			}

			if logSize == 0 || game.potentialRound.WPAlog[logSize-1].tick+(game.tickRate) < p.GameState().IngameTick() {
				newWPAentry := &wpalog{
					round:               int(game.potentialRound.roundNum),
					clock:               clock,
					planted:             planted,
					tick:                p.GameState().IngameTick(),
					ctAlive:             game.flags.ctAlive,
					tAlive:              game.flags.tAlive,
					ctEquipVal:          calculateTeamEquipmentValue(game, p.GameState().TeamCounterTerrorists(), p),
					tEquipVal:           calculateTeamEquipmentValue(game, p.GameState().TeamTerrorists(), p),
					ctFlashes:           calculateTeamEquipmentNum(game, p.GameState().TeamCounterTerrorists(), 15, p),
					ctSmokes:            calculateTeamEquipmentNum(game, p.GameState().TeamCounterTerrorists(), 16, p),
					ctMolys:             calculateTeamEquipmentNum(game, p.GameState().TeamCounterTerrorists(), 17, p),
					ctFrags:             calculateTeamEquipmentNum(game, p.GameState().TeamCounterTerrorists(), 14, p),
					tFlashes:            calculateTeamEquipmentNum(game, p.GameState().TeamTerrorists(), 15, p),
					tSmokes:             calculateTeamEquipmentNum(game, p.GameState().TeamTerrorists(), 16, p),
					tMolys:              calculateTeamEquipmentNum(game, p.GameState().TeamTerrorists(), 17, p),
					tFrags:              calculateTeamEquipmentNum(game, p.GameState().TeamTerrorists(), 14, p),
					closestCTDisttoBomb: closestCTDisttoBomb(game, p.GameState().TeamCounterTerrorists(), p.GameState().Bomb(), p),
					kits:                numOfKits(game, p.GameState().TeamCounterTerrorists(), p),
					ctArmor:             playersWithArmor(game, p.GameState().TeamCounterTerrorists(), p),
					tArmor:              playersWithArmor(game, p.GameState().TeamTerrorists(), p),
				}
				game.potentialRound.WPAlog = append(game.potentialRound.WPAlog, newWPAentry)
			}
		}

		if game.flags.inRound && game.flags.lastTickProcessed+(4*game.tickRate) < p.GameState().IngameTick() {
			game.flags.lastTickProcessed = p.GameState().IngameTick()
			game.flags.ticksProcessed += 1

			//this will be triggered every 4 seconds of in round time after the first 10 seconds

			//check for lurker
			if game.flags.tAlive > 2 && !game.flags.postWinCon && p.GameState().IngameTick() > (18*game.tickRate)+game.potentialRound.startingTick {
				membersT := getTeamMembers(p.GameState().TeamTerrorists(), game, p)
				for _, terrorist := range membersT {
					if terrorist.IsAlive() {
						for _, teammate := range membersT {
							if terrorist.SteamID64 != teammate.SteamID64 && teammate.IsAlive() {
								dist := int(terrorist.Position().Distance(teammate.Position()))
								if dist < 500 {
									//invalidate the lurk blip b/c we have a close teammate
									game.potentialRound.playerStats[terrorist.SteamID64].distanceToTeammates = -999999
								}
								if game.potentialRound.playerStats[terrorist.SteamID64] != nil {
									game.potentialRound.playerStats[terrorist.SteamID64].distanceToTeammates += dist
								} else {
									fmt.Println("THIS IS WHERE WE BROKE_______________________________---------------------------------------------------")
								}
							}
						}
					}
				}
				var lurkerSteam uint64
				lurkerDist := 999999
				for _, terrorist := range membersT {
					if terrorist.IsAlive() {
						if game.potentialRound.playerStats[terrorist.SteamID64] == nil {
							if DEBUG {
								fmt.Println(terrorist.Name)
							}
						}
						dist := game.potentialRound.playerStats[terrorist.SteamID64].distanceToTeammates
						if dist < lurkerDist && dist > 0 {
							lurkerDist = dist
							lurkerSteam = terrorist.SteamID64
						}
					}
				}
				if lurkerSteam != 0 {
					game.potentialRound.playerStats[lurkerSteam].lurkerBlips += 1
				}
			}
		}
	})

	p.RegisterEventHandler(func(e events.RoundStart) {
		if DEBUG {
			fmt.Println("Round Start", p.GameState().TotalRoundsPlayed())
		}

		game.flags.roundStartedAt = p.GameState().IngameTick()

	})

	p.RegisterEventHandler(func(e events.RoundFreezetimeEnd) {
		if DEBUG {
			fmt.Printf("Round Freeze Time End\n")
		}
		pistol := false

		//we are going to check to see if the first pistol is actually starting
		membersT := getTeamMembers(p.GameState().TeamTerrorists(), game, p)
		membersCT := getTeamMembers(p.GameState().TeamCounterTerrorists(), game, p)
		if len(membersT) != 0 && len(membersCT) != 0 {
			if membersT[0].Money()+membersT[0].MoneySpentThisRound() == 800 && membersCT[0].Money()+membersCT[0].MoneySpentThisRound() == 800 {
				//start the game
				if !game.flags.hasGameStarted {
					initGameStart()
				}

				//track the pistol
				pistol = true
			} else if membersT[0].Money()+membersT[0].MoneySpentThisRound() == 0 && membersCT[0].Money()+membersCT[0].MoneySpentThisRound() == 0 {
				game.potentialRound.knifeRound = true
				if DEBUG {
					fmt.Println("------------KNIFEROUND-----------")
				}
				game.flags.hasGameStarted = false
			}
		}
		if DEBUG {
			fmt.Println("Has the Game Started?", game.flags.hasGameStarted)
		}

		if game.flags.isGameLive {
			//init round stats
			game.flags.inRound = true
			initRound()
			if pistol {
				for _, team := range game.potentialRound.teamStats {
					team.pistols = 1
				}
			}

		}

	})

	p.RegisterEventHandler(func(e events.RoundEnd) {
		if game.flags.isGameLive {

			game.flags.didRoundEndFire = true
			if DEBUG {
				fmt.Println("Round", p.GameState().TotalRoundsPlayed(), "End", e.WinnerState.ClanName(), "won", "this determined from e.WinnerState.ClanName()")

				fmt.Println("e.WinnerState.ID()", e.WinnerState.ID(), "and", "e.Winner", e.Winner, "and", "e.WinnerState.Team()", e.WinnerState.Team())
			}
			validWinner := true
			if e.Winner < 2 {
				validWinner = false
				//and set the integrity flag to false

			} else if e.Winner == 2 {
				game.flags.tMoney = true
			} else {
				//we need to check if the game is over

			}

			//we want to actually process the round
			//TODO: VERIFY BEHAVIOR OF THIS. I REMOVED +1 FROM TOTALROUNDSPLAYED
			if game.flags.isGameLive && validWinner && game.flags.roundIntegrityStart == p.GameState().TotalRoundsPlayed() {
				game.potentialRound.winnerENUM = int(e.Winner)
				processRoundOnWinCon(validateTeamName(game, e.WinnerState.ClanName(), e.WinnerState.Team()))

				//check last round
				roundWinnerScore := game.teams[validateTeamName(game, e.WinnerState.ClanName(), e.WinnerState.Team())].score
				roundLoserScore := game.teams[validateTeamName(game, e.LoserState.ClanName(), e.LoserState.Team())].score
				if DEBUG {
					fmt.Println("winner Rounds", roundWinnerScore)
					fmt.Println("loser Rounds", roundLoserScore)
				}

				if game.roundsToWin == 16 {
					//check for normal win
					if roundWinnerScore == 16 && roundLoserScore < 15 {
						//normal win
						game.winnerClanName = game.potentialRound.winnerClanName
						processRoundFinal(true)
					} else if roundWinnerScore > 15 { //check for OT win
						overtime := ((roundWinnerScore+roundLoserScore)-30-1)/6 + 1
						//OT win
						if (roundWinnerScore-15-1)/3 == overtime {
							game.winnerClanName = game.potentialRound.winnerClanName
							processRoundFinal(true)
						}
					}
				} else if game.roundsToWin == 9 {
					//check for normal win
					if roundWinnerScore == 9 && roundLoserScore < 8 {
						//normal win
						game.winnerClanName = game.potentialRound.winnerClanName
						processRoundFinal(true)
					} else if roundWinnerScore == 8 && roundLoserScore == 8 { //check for tie
						//tie
						game.winnerClanName = game.potentialRound.winnerClanName
						processRoundFinal(true)
					}
				} else if game.roundsToWin == 13 {
					//check for normal win
					if roundWinnerScore == 13 && roundLoserScore < 12 {
						//normal win
						game.winnerClanName = game.potentialRound.winnerClanName
						processRoundFinal(true)
					} else if roundWinnerScore > 12 { //check for OT win
						overtime := ((roundWinnerScore+roundLoserScore)-24-1)/6 + 1
						//OT win
						if (roundWinnerScore-12-1)/3 == overtime {
							game.winnerClanName = game.potentialRound.winnerClanName
							processRoundFinal(true)
						}
					}
				}
			}

			//check last round
			//or check overtime win

		}
	})

	//round end official doesnt fire on the last round
	p.RegisterEventHandler(func(e events.ScoreUpdated) {
		//CS2 swapped this event to be before RoundEnd
		//We have relied on this as a back up for failed RoundEnd events
		//may revisit depending on event reliability
	})

	//round end official doesnt fire on the last round
	p.RegisterEventHandler(func(e events.RoundEndOfficial) {

		if DEBUG {
			fmt.Printf("Round End Official\n")
		}

		if !game.flags.didRoundEndFire {
			game.flags.roundIntegrityEnd -= 1
		}

		if DEBUG {
			fmt.Println("isGameLive", game.flags.isGameLive, "roundIntegrityEnd", game.flags.roundIntegrityEnd, "pTotalRoundsPlayed", p.GameState().TotalRoundsPlayed())
		}

		if game.flags.isGameLive && game.flags.roundIntegrityEnd == p.GameState().TotalRoundsPlayed() {
			processRoundFinal(false)
		}
	})

	// Register handler on kill events
	p.RegisterEventHandler(func(e events.Kill) {
		flashAssister := ""
		if game.flags.isGameLive && isDuringExpectedRound(game, p) {
			pS := game.potentialRound.playerStats
			tick := p.GameState().IngameTick()

			killerExists := false
			victimExists := false
			assisterExists := false
			if e.Killer != nil && pS[e.Killer.SteamID64] != nil {
				killerExists = true
			}
			if e.Victim != nil && pS[e.Victim.SteamID64] != nil {
				victimExists = true
			}
			if e.Assister != nil && pS[e.Assister.SteamID64] != nil {
				assisterExists = true
			}

			killValue := 1.0
			multiplier := 1.0
			traded := false
			assisted := false
			flashAssisted := false

			//death logic (traded here)
			if victimExists {
				pS[e.Victim.SteamID64].deaths += 1
				pS[e.Victim.SteamID64].deathTick = tick
				if e.Victim.Team == 2 {
					game.flags.tAlive -= 1
					pS[e.Victim.SteamID64].deathPlacement = float64(game.potentialRound.initTerroristCount - game.flags.tAlive)
					//pS[e.Victim.SteamID64].tADP = float64(game.potentialRound.initTerroristCount - game.flags.tAlive)
				} else if e.Victim.Team == 3 {
					game.flags.ctAlive -= 1
					pS[e.Victim.SteamID64].deathPlacement = float64(game.potentialRound.initCTerroristCount - game.flags.ctAlive)
					//pS[e.Victim.SteamID64].ctADP = float64(game.potentialRound.initCTerroristCount - game.flags.ctAlive)
				} else {
					//else log an error
				}

				//do 4v5 calc
				if game.flags.openingKill && game.potentialRound.initCTerroristCount+game.potentialRound.initTerroristCount == 10 {
					//the 10th player died
					_4v5Team := pS[e.Victim.SteamID64].teamClanName
					game.potentialRound.teamStats[_4v5Team]._4v5s = 1
					for teamName, team := range game.potentialRound.teamStats {
						if teamName != _4v5Team {
							team._5v4s = 1
						}
					}
				}

				//add support damage
				for suppSteam, suppDMG := range pS[e.Victim.SteamID64].damageList {
					if killerExists && suppSteam != e.Killer.SteamID64 {
						pS[suppSteam].suppDamage += suppDMG
						if pS[suppSteam].suppDamage > 60 {
							pS[suppSteam].suppRounds = 1
						}
					} else if !killerExists {
						pS[suppSteam].suppDamage += suppDMG
						if pS[suppSteam].suppDamage > 60 {
							pS[suppSteam].suppRounds = 1
						}
					}

				}

				//check clutch start

				if !game.flags.postWinCon {
					if game.flags.tAlive == 1 && game.flags.tClutchVal == 0 {
						game.flags.tClutchVal = game.flags.ctAlive
						membersT := getTeamMembers(p.GameState().TeamTerrorists(), game, p)
						for _, terrorist := range membersT {
							if terrorist.IsAlive() && e.Victim.SteamID64 != terrorist.SteamID64 {
								game.flags.tClutchSteam = terrorist.SteamID64
								if DEBUG {
									fmt.Println("Clutch opportunity:", terrorist.Name, game.flags.tClutchVal)
								}
							}
						}
					}
					if game.flags.ctAlive == 1 && game.flags.ctClutchVal == 0 {
						game.flags.ctClutchVal = game.flags.tAlive
						membersCT := getTeamMembers(p.GameState().TeamCounterTerrorists(), game, p)
						for _, counterTerrorist := range membersCT {
							if counterTerrorist.IsAlive() && e.Victim.SteamID64 != counterTerrorist.SteamID64 {
								game.flags.ctClutchSteam = counterTerrorist.SteamID64
								if DEBUG {
									fmt.Println("Clutch opportunity:", counterTerrorist.Name, game.flags.ctClutchVal)
								}
							}
						}
					}
				}

				pS[e.Victim.SteamID64].ticksAlive = tick - game.potentialRound.startingTick
				for deadGuySteam, deadTick := range (*game.potentialRound).playerStats[e.Victim.SteamID64].tradeList {
					if tick-deadTick < tradeCutoff*game.tickRate {
						pS[deadGuySteam].traded = 1
						pS[deadGuySteam].eac += 1
						pS[deadGuySteam].kastRounds = 1
					}
				}
			}

			//assist logic
			if assisterExists && victimExists && e.Assister.TeamState.ID() != e.Victim.TeamState.ID() {
				//this logic needs to be replaced -yeti does not remember why he wrote this
				pS[e.Assister.SteamID64].assists += 1
				pS[e.Assister.SteamID64].eac += 1
				pS[e.Assister.SteamID64].kastRounds = 1
				pS[e.Assister.SteamID64].suppRounds = 1
				assisted = true
				if e.AssistedFlash {
					pS[e.Assister.SteamID64].fAss += 1
					flashAssisted = true
					flashAssister = e.Assister.Name
					//fmt.Println("VALVE FLASH ASSIST")
				} else if float64(p.GameState().IngameTick()) < pS[e.Victim.SteamID64].mostRecentFlashVal {
					//this will trigger if there is both a flash assist and a damage assist
					pS[pS[e.Victim.SteamID64].mostRecentFlasher].fAss += 1
					pS[pS[e.Victim.SteamID64].mostRecentFlasher].eac += 1
					pS[pS[e.Victim.SteamID64].mostRecentFlasher].suppRounds = 1
					flashAssisted = true
					flashAssister = pS[pS[e.Victim.SteamID64].mostRecentFlasher].name
				}

			}

			//kill logic (trades here)
			if killerExists && victimExists && e.Killer.TeamState.ID() != e.Victim.TeamState.ID() {
				pS[e.Killer.SteamID64].kills += 1
				pS[e.Killer.SteamID64].kastRounds = 1
				pS[e.Killer.SteamID64].rwk = 1
				pS[e.Killer.SteamID64].tradeList[e.Victim.SteamID64] = tick
				if e.Weapon.Type == 309 {
					pS[e.Killer.SteamID64].awpKills += 1
					if e.Killer.Team == 3 {
						pS[e.Killer.SteamID64].ctAWP += 1
					}
				}
				if e.IsHeadshot {
					pS[e.Killer.SteamID64].hs += 1
				}
				for _, deadTick := range (*game.potentialRound).playerStats[e.Victim.SteamID64].tradeList {
					if tick-deadTick < tradeCutoff*game.tickRate {
						pS[e.Killer.SteamID64].trades += 1
						traded = true
						break
					}
				}

				killerTeam := e.Killer.Team
				if game.flags.prePlant {
					//normal base value
					if killerTeam == 2 {
						//taking site by T
						killValue = 1.2
					} else if killerTeam == 3 {
						//site Defense by CT
						killValue = 1
					}
				} else if game.flags.postPlant {
					//site D or retake
					if killerTeam == 2 {
						//site Defense by T
						killValue = 1
					} else if killerTeam == 3 {
						//retake
						killValue = 1.2
					}
				} else if game.flags.postWinCon {
					//exit or chase
					if game.potentialRound.winnerENUM == 2 { //Ts win
						if killerTeam == 2 { //chase
							killValue = 0.8
						}
						if killerTeam == 3 { //exit
							killValue = 0.6
						}
					} else if game.potentialRound.winnerENUM == 3 { //CTs win
						if killerTeam == 2 { //T kill in lost round
							killValue = 0.5
						}
						if killerTeam == 3 { //CT kill in won round
							if game.flags.tMoney {
								killValue = 0.6
							} else {
								killValue = 0.8
							}
						}
					}
				}

				if game.flags.openingKill {
					game.flags.openingKill = false

					pS[e.Killer.SteamID64].ok = 1
					pS[e.Victim.SteamID64].ol = 1

					if killerTeam == 2 { //T entry/opener {
						if game.flags.prePlant {
							multiplier += 0.8
							pS[e.Killer.SteamID64].entries = 1
						} else {
							multiplier += 0.3
						}
					} else if killerTeam == 3 { //CT opener
						multiplier += 0.5
					}

				} else if traded {
					multiplier += 0.3
				}

				if flashAssisted { //flash assisted kill
					multiplier += 0.2
				}
				if assisted { //assisted kill
					killValue -= 0.15
					pS[e.Assister.SteamID64].impactPoints += 0.15
				}

				killValue *= multiplier

				ecoRatio := float64(e.Victim.EquipmentValueCurrent()) / float64(e.Killer.EquipmentValueCurrent())
				ecoMod := 1.0
				if ecoRatio > 4 {
					ecoMod += 0.25
				} else if ecoRatio > 2 {
					ecoMod += 0.14
				} else if ecoRatio < 0.25 {
					ecoMod -= 0.25
				} else if ecoRatio < 0.5 {
					ecoMod -= 0.14
				}
				killValue *= ecoMod

				pS[e.Killer.SteamID64].killPoints += killValue
			}

		}
		var hs string
		if e.IsHeadshot {
			hs = " (HS)"
		}
		var wallBang string
		if e.PenetratedObjects > 0 {
			wallBang = " (WB)"
		}
		if DEBUG {
			fmt.Printf("%d %s <%v%s%s> %s at %d flash assist by %s\n", e.Victim.Team, e.Killer, e.Weapon, hs, wallBang, e.Victim, p.GameState().IngameTick(), flashAssister)
		}
	})

	p.RegisterEventHandler(func(e events.PlayerHurt) {
		//fmt.Printf("Player Hurt\n")
		if game.flags.isGameLive {
			equipment := e.Weapon.Type
			if e.Player != nil {
				game.potentialRound.playerStats[e.Player.SteamID64].damageTaken += e.HealthDamageTaken
			}
			if e.Player != nil && e.Attacker != nil && e.Player.Team != e.Attacker.Team {
				game.potentialRound.playerStats[e.Attacker.SteamID64].damage += e.HealthDamageTaken

				//add to damage list for supp damage calc
				game.potentialRound.playerStats[e.Player.SteamID64].damageList[e.Attacker.SteamID64] += e.HealthDamageTaken

				if equipment >= 500 && equipment <= 506 {
					game.potentialRound.playerStats[e.Attacker.SteamID64].utilDmg += e.HealthDamageTaken
					if equipment == 506 {
						game.potentialRound.playerStats[e.Attacker.SteamID64].nadeDmg += e.HealthDamageTaken
					}
					if equipment == 502 || equipment == 503 {
						game.potentialRound.playerStats[e.Attacker.SteamID64].infernoDmg += e.HealthDamageTaken
					}
				}
			}
		}
	})

	p.RegisterEventHandler(func(e events.PlayerFlashed) {
		//fmt.Println("Player Flashed")
		if game.flags.isGameLive && e.Player != nil && e.Attacker != nil {
			tick := float64(p.GameState().IngameTick())
			blindTicks := e.FlashDuration().Seconds() * 128.0
			victim := e.Player
			flasher := e.Attacker
			if flasher.Team != victim.Team && blindTicks > 128.0 && victim.IsAlive() && (float64(victim.FlashDuration) < (blindTicks/128.0 + 1)) {
				game.potentialRound.playerStats[flasher.SteamID64].ef += 1
				game.potentialRound.playerStats[flasher.SteamID64].enemyFlashTime += (blindTicks / 128.0)
				if tick+blindTicks > game.potentialRound.playerStats[victim.SteamID64].mostRecentFlashVal {
					game.potentialRound.playerStats[victim.SteamID64].mostRecentFlashVal = tick + blindTicks
					game.potentialRound.playerStats[victim.SteamID64].mostRecentFlasher = flasher.SteamID64
				}

			}
			if flasher.Name != "" && printDebugLog {
				debugMsg := fmt.Sprintf("%s flashed %s for %.2f at %d. He was %f blind.\n", flasher, victim, blindTicks/128, int(tick), victim.FlashDuration)
				debugFile.WriteString(debugMsg)
				debugFile.Sync()
			}

		}
		//fmt.Println("Player Flashed", blindTicks, e.Attacker)
	})

	p.RegisterEventHandler(func(e events.RoundImpactScoreData) {
		if DEBUG {
			fmt.Println("-------ROUNDIMPACTSCOREDATA", e.RawMessage)
		}
	})

	p.RegisterEventHandler(func(e events.BombPlanted) {
		if DEBUG {
			fmt.Printf("Bomb Planted\n")
		}
		if game.flags.isGameLive && !game.flags.postWinCon {
			game.flags.prePlant = false
			game.flags.postPlant = true
			game.flags.tMoney = true
			game.potentialRound.planter = e.BombEvent.Player.SteamID64
			game.potentialRound.bombStartTick = p.GameState().IngameTick()
		}
	})

	p.RegisterEventHandler(func(e events.BombDefused) {
		if DEBUG {
			fmt.Println("Bomb Defused by", e.BombEvent.Player.Name)
		}
		if game.flags.isGameLive && !game.flags.postWinCon {
			game.flags.prePlant = false
			game.flags.postPlant = false
			game.flags.postWinCon = true
			game.potentialRound.endDueToBombEvent = true
			game.potentialRound.defuser = e.Player.SteamID64
			game.potentialRound.playerStats[e.BombEvent.Player.SteamID64].impactPoints += 0.5
		}
	})

	p.RegisterEventHandler(func(e events.BombExplode) {
		if DEBUG {
			fmt.Printf("Bomb Exploded\n")
		}
		if game.flags.isGameLive && !game.flags.postWinCon {
			game.flags.prePlant = false
			game.flags.postPlant = false
			game.flags.postWinCon = true
			game.potentialRound.endDueToBombEvent = true
			if game.potentialRound.planter != 0 {
				game.potentialRound.playerStats[game.potentialRound.planter].impactPoints += 0.5
			}
		}
	})

	p.RegisterEventHandler(func(e events.GrenadeProjectileThrow) {
		//fmt.Println("Grenade Thrown", e.Projectile.WeaponInstance.Type)
		if game.flags.isGameLive {
			if e.Projectile.WeaponInstance.Type == 506 {
				game.potentialRound.playerStats[e.Projectile.Thrower.SteamID64].nadesThrown += 1
			} else if e.Projectile.WeaponInstance.Type == 505 {
				game.potentialRound.playerStats[e.Projectile.Thrower.SteamID64].smokeThrown += 1
			} else if e.Projectile.WeaponInstance.Type == 504 {
				game.potentialRound.playerStats[e.Projectile.Thrower.SteamID64].flashThrown += 1
			} else if e.Projectile.WeaponInstance.Type == 502 || e.Projectile.WeaponInstance.Type == 503 {
				game.potentialRound.playerStats[e.Projectile.Thrower.SteamID64].firesThrown += 1
			}

		}
	})

	p.RegisterEventHandler(func(e events.PlayerTeamChange) {
		if DEBUG {
			fmt.Println("Player Changed Team:", e.Player, e.OldTeam, e.NewTeam)
		}

		if game.flags.isGameLive && game.flags.inRound {
			if e.NewTeam > 1 {
				//we are joining an actual team
				if game.potentialRound.playerStats[e.Player.SteamID64] == nil && e.Player.IsBot && e.Player.IsAlive() {
					//get team
					team := e.NewTeamState
					player := &playerStats{name: e.Player.Name, steamID: e.Player.SteamID64, isBot: e.Player.IsBot, side: int(team.Team()), teamENUM: team.ID(), teamClanName: validateTeamName(game, team.ClanName(), team.Team()), health: 100, tradeList: make(map[uint64]int), damageList: make(map[uint64]int)}
					game.potentialRound.playerStats[player.steamID] = player
				}
			}
		}
	})

	p.RegisterEventHandler(func(e events.PlayerDisconnected) {
		if DEBUG {
			fmt.Println("Player DC", e.Player)
			if game.reconnectedPlayers[e.Player.SteamID64] {
				game.reconnectedPlayers[e.Player.SteamID64] = false
			}
		}

		//update alive players
		if game.flags.isGameLive {
			game.flags.tAlive = 0
			game.flags.ctAlive = 0

			membersT := getTeamMembers(p.GameState().TeamTerrorists(), game, p)
			for _, terrorist := range membersT {
				if terrorist.IsAlive() {
					game.flags.tAlive += 1
				}
			}
			membersCT := getTeamMembers(p.GameState().TeamCounterTerrorists(), game, p)
			for _, counterTerrorist := range membersCT {
				if counterTerrorist.IsAlive() {
					game.flags.ctAlive += 1
				}
			}
		}

	})

	p.RegisterEventHandler(func(e events.Footstep) {
		if game.flags.isGameLive {
			game.flags.inRound = true
		}

	})

	if printChatLog {
		p.RegisterEventHandler(func(e events.ChatMessage) {
			chatMsg := fmt.Sprintf("%s: \"%s\" at %d\n", e.Sender, e.Text, p.GameState().IngameTick())
			if printChatLog {
				chatFile.WriteString(chatMsg)
				chatFile.Sync()
			}
			//check(err)
		})
	}

	// Parse to end
	err = p.ParseToEnd()
	if err != nil {
		if err.Error() == "demo stream ended unexpectedly (ErrUnexpectedEndOfDemo)" {
			fmt.Println("Ignoring ErrUnexpectedEndOfDemo to continue processing demo.")
		} else {
			panic(err)
		}
	}

	//----END OF MATCH PROCESSING----
	//we want to iterate through rounds backwards to make sure their are no repeats

	endOfMatchProcessing(game)
	beginOutput(game)
	if DEBUG {
		fmt.Println("coreID: ", game.coreID)
	}

	if BACKEND_PUSHING && game.coreID != "" {
		token := authenticate()
		if verifyOriginalMatch(token, game.coreID, game.mapNum) || FORCE_NEW_STATS_UPLOAD { //
			addMatchStats(token, game)
		}
	} else if game.coreID != "" && DEBUG {
		fmt.Println("This is not a CSC demo D:")
	}

	fmt.Println("Demo", game.mid, " is complete!")
	defer swg.Done()
	//cleanup()

}
