package main

import (
	//"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	dem "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
	events "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/events"
)

//TODO
//"Catch up on the score"

const DEBUG = false

//globals
const printChatLog = true
const printDebugLog = true

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

type game struct {
	rounds           []*round
	teams            map[int]*team
	flags            flag
	mapName          string
	tickRate         int
	tickLength       uint64
	roundsToWin      int //30 or 16
	totalPlayerStats map[uint64]*playerStats
}

type flag struct {
	//all our sentinals and shit
	hasGameStarted            bool
	isGameLive                bool
	isGameOver                bool
	prePlant                  bool
	postPlant                 bool
	postWinCon                bool
	roundIntegrityStart       int
	roundIntegrityEnd         int
	roundIntegrityEndOfficial int

	//for the round (gets reset on a new round) maybe should be in a new struct
	tAlive        int
	ctAlive       int
	tMoney        bool
	tClutchVal    int
	ctClutchVal   int
	tClutchSteam  uint64
	ctClutchSteam uint64
	openingKill   bool
}

type team struct {
	id    int
	name  string
	score uint8
}

type teamStats struct {
	winPoints    float64
	impactPoints float64
	_4v5         int
	_5v4         int
}

type round struct {
	//round value
	roundNum            int8
	startingTick        int
	endingTick          int
	playerStats         map[uint64]*playerStats
	teamStats           map[int]*teamStats
	initTerroristCount  int
	initCTerroristCount int
	winnerID            int //this is the unique ID which should not change
	winnerENUM          int //this effectively represents the side that won: 2 (T) or 3 (CT)
	integrityCheck      bool
	planter             uint64
	defuser             uint64
}

type playerStats struct {
	name    string
	steamID uint64
	teamID  int
	rounds  int
	//playerPoints float32
	//teamPoints float32
	damage         int
	rating         float64
	kills          uint8
	assists        int8
	deaths         int8
	deathTick      int
	deathPlacement float64
	ticksAlive     int
	trades         int
	traded         int
	ok             int
	ol             int
	cl_1           int
	cl_2           int
	cl_3           int
	cl_4           int
	cl_5           int
	_2k            int
	_3k            int
	_4k            int
	_5k            int
	nadeDmg        int
	infernoDmg     int
	utilDmg        int
	ef             int
	fAss           int
	enemyFlashTime float64
	hs             int
	kastRounds     float64
	saves          int
	entries        int
	killPoints     float64
	impactPoints   float64
	winPoints      float64
	awpKills       int

	//derived
	atd          int
	kast         float64
	killPointAvg float64

	//"flags"
	health             int
	tradeList          map[uint64]int
	mostRecentFlasher  uint64
	mostRecentFlashVal float64
}

func main() {

	//for concurrency
	//for i := 0; i < 10; i++ {
	//	go processDemo("faceit1.dem")
	//}

	processDemo("faceit1.dem")

	var input string
	fmt.Scanln(&input)
}

func initGameObject() *game {
	g := game{}
	g.rounds = make([]*round, 0)

	g.flags.hasGameStarted = false
	g.flags.isGameLive = false
	g.flags.isGameOver = false
	g.flags.prePlant = true
	g.flags.postPlant = false
	g.flags.postWinCon = false
	//these three vars to check if we have a complete round
	g.flags.roundIntegrityStart = -1
	g.flags.roundIntegrityEnd = -1
	g.flags.roundIntegrityEndOfficial = -1

	return &g
}

func processDemo(demoName string) {

	game := initGameObject()

	/*
	   var players = new Array();
	   var openingKill = true;
	   var planter;
	   var errLog = "";

	   var debugLog = [];
	*/

	f, err := os.Open(demoName)
	//f, err := os.Open("league1.dem")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	p := dem.NewParser(f)
	defer p.Close()

	//must parse header to get header info
	header, err := p.ParseHeader()
	if err != nil {
		panic(err)
	}

	//set map name
	game.mapName = strings.Title((header.MapName)[3:])
	fmt.Println("Map is", game.mapName)

	//set tick rate
	game.tickRate = int(math.Round(p.TickRate()))
	fmt.Println("Tick rate is", game.tickRate)

	//creating a file to dump chat log into
	fmt.Printf("Creating chatLog file.\n")
	os.Mkdir("out", 0777)
	chatFile, chatFileErr := os.Create("out/chatLog.txt")
	if chatFileErr != nil {
		fmt.Printf("Error in opening chatLog file.\n")
	}
	defer chatFile.Close()

	fmt.Printf("Creating debug file.\n")
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

		game.teams = make(map[int]*team)

		teamTemp := p.GameState().TeamTerrorists()
		game.teams[teamTemp.ID()] = &team{id: teamTemp.ID(), name: teamTemp.ClanName()}
		teamTemp = p.GameState().TeamCounterTerrorists()
		game.teams[teamTemp.ID()] = &team{id: teamTemp.ID(), name: teamTemp.ClanName()}

		//to handle short and long matches
		if p.GameState().ConVars()["mp_maxrounds"] != "30" {
			maxRounds, fuckOFF := strconv.Atoi(p.GameState().ConVars()["mp_maxrounds"])
			if fuckOFF == nil {
				game.roundsToWin = maxRounds/2 + 1
			} else {
				//ADD TO ERROR LOG
				game.roundsToWin = maxRounds/2 + 1
				//maybe this gives us a way to check for short vs long match
			}
		} else {
			game.roundsToWin = 16 //we will assume long match in case convar is not set
		}

	}

	initRound := func() {
		game.flags.roundIntegrityStart = p.GameState().TotalRoundsPlayed() + 1
		fmt.Println("We are starting round", game.flags.roundIntegrityStart)
		game.flags.tMoney = false
		game.flags.openingKill = true

		currRoundObj := &round{roundNum: int8(game.flags.roundIntegrityStart), startingTick: p.GameState().IngameTick()}
		currRoundObj.playerStats = make(map[uint64]*playerStats)
		currRoundObj.teamStats = make(map[int]*teamStats)

		//set players in playerStats for the round
		terrorists := p.GameState().TeamTerrorists()
		for _, terrorist := range terrorists.Members() {
			player := &playerStats{name: terrorist.Name, steamID: terrorist.SteamID64, teamID: terrorists.ID(), health: 100, tradeList: make(map[uint64]int)}
			currRoundObj.playerStats[player.steamID] = player
		}
		counterTerrorists := p.GameState().TeamCounterTerrorists()
		for _, counterTerrorist := range counterTerrorists.Members() {
			player := &playerStats{name: counterTerrorist.Name, steamID: counterTerrorist.SteamID64, teamID: counterTerrorists.ID(), health: 100, tradeList: make(map[uint64]int)}
			currRoundObj.playerStats[player.steamID] = player
		}

		//set teams in teamStats for the round
		currRoundObj.teamStats[p.GameState().TeamTerrorists().ID()] = &teamStats{}
		currRoundObj.teamStats[p.GameState().TeamCounterTerrorists().ID()] = &teamStats{}

		if len(game.rounds) == 0 {
			game.rounds = append(game.rounds, currRoundObj)
		} else {
			game.rounds[0] = currRoundObj
		}

		// if len(game.rounds) < game.flags.roundIntegrityStart {
		// 	//game.rounds[0] = currRoundObj
		// } else {
		// 	//game.rounds[roundIntegrityStart - 1] = currRoundObj
		// 	fmt.Println(game.rounds[game.flags.roundIntegrityStart - 1].integrityCheck)
		// 	//game.rounds = append(game.rounds, currRoundObj)
		// }

		//track the number of people alive for clutch checking and record keeping
		game.flags.tAlive = len(terrorists.Members())
		game.flags.ctAlive = len(counterTerrorists.Members())
		game.rounds[0].initTerroristCount = game.flags.tAlive
		game.rounds[0].initCTerroristCount = game.flags.ctAlive

		//reset various flags
		game.flags.prePlant = true
		game.flags.postPlant = false
		game.flags.postWinCon = false
		game.flags.tClutchVal = 0
		game.flags.ctClutchVal = 0
		game.flags.tClutchSteam = 0
		game.flags.ctClutchSteam = 0
	}

	processRoundOnWinCon := func(winnerID int) {
		game.flags.roundIntegrityEnd = p.GameState().TotalRoundsPlayed() + 1
		fmt.Println("We are processing round win con stuff", game.flags.roundIntegrityEnd)

		game.flags.prePlant = false
		game.flags.postPlant = false
		game.flags.postWinCon = true

		//set winner
		game.rounds[0].winnerID = winnerID
		game.teams[game.rounds[0].winnerID].score += 1
		//check clutch

	}

	processRoundFinal := func(lastRound bool) {
		game.rounds[0].endingTick = p.GameState().IngameTick()
		game.flags.roundIntegrityEndOfficial = p.GameState().TotalRoundsPlayed()
		if lastRound {
			game.flags.roundIntegrityEndOfficial += 1
		}
		fmt.Println("We are processing round final stuff", game.flags.roundIntegrityEndOfficial)
		fmt.Println(len(game.rounds))

		//we have the entire round uninterrupted
		if game.flags.roundIntegrityStart == game.flags.roundIntegrityEnd && game.flags.roundIntegrityEnd == game.flags.roundIntegrityEndOfficial {
			game.rounds[0].integrityCheck = true

			//check saves

			//set the clutch
			if game.rounds[0].winnerENUM == 2 && game.flags.tClutchSteam != 0 {
				game.rounds[0].playerStats[game.flags.tClutchSteam].impactPoints += clutchBonus[game.flags.tClutchVal]
				switch game.flags.tClutchVal {
				case 1:
					game.rounds[0].playerStats[game.flags.tClutchSteam].cl_1 = 1
				case 2:
					game.rounds[0].playerStats[game.flags.tClutchSteam].cl_2 = 1
				case 3:
					game.rounds[0].playerStats[game.flags.tClutchSteam].cl_3 = 1
				case 4:
					game.rounds[0].playerStats[game.flags.tClutchSteam].cl_4 = 1
				case 5:
					game.rounds[0].playerStats[game.flags.tClutchSteam].cl_5 = 1
				}
			} else if game.rounds[0].winnerENUM == 3 && game.flags.ctClutchSteam != 0 {
				switch game.flags.ctClutchVal {
				case 1:
					game.rounds[0].playerStats[game.flags.ctClutchSteam].cl_1 = 1
				case 2:
					game.rounds[0].playerStats[game.flags.ctClutchSteam].cl_2 = 1
				case 3:
					game.rounds[0].playerStats[game.flags.ctClutchSteam].cl_3 = 1
				case 4:
					game.rounds[0].playerStats[game.flags.ctClutchSteam].cl_4 = 1
				case 5:
					game.rounds[0].playerStats[game.flags.ctClutchSteam].cl_5 = 1
				}
			}

			//add multikills & saves & misc
			for _, player := range (game.rounds[0]).playerStats {
				if player.deaths == 0 {
					player.kastRounds = 1
					if player.teamID != game.rounds[0].winnerID {
						fmt.Println("Player on team this saved:", player.teamID)
						player.saves = 1
					}
				}
				game.rounds[0].playerStats[player.steamID].impactPoints += player.killPoints
				game.rounds[0].playerStats[player.steamID].impactPoints += float64(player.damage) / float64(250)
				game.rounds[0].playerStats[player.steamID].impactPoints += multikillBonus[player.kills]
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

				if player.teamID == game.rounds[0].winnerENUM {
					player.winPoints = player.impactPoints
				}
			}

			//add our valid round
			game.rounds = append(game.rounds, game.rounds[0])
		}

		//endRound function functionality

	}

	//-------------ALL OUR EVENTS---------------------

	p.RegisterEventHandler(func(e events.RoundStart) {
		fmt.Printf("Round Start\n")
	})

	p.RegisterEventHandler(func(e events.RoundFreezetimeEnd) {
		fmt.Printf("Round Freeze Time End\n")

		//we are going to check to see if the first pistol is actually starting
		membersT := p.GameState().TeamTerrorists().Members()
		membersCT := p.GameState().TeamCounterTerrorists().Members()
		if len(membersT) != 0 && len(membersCT) != 0 {
			if membersT[0].Money()+membersT[0].MoneySpentThisRound() == 800 && membersCT[0].Money()+membersCT[0].MoneySpentThisRound() == 800 {
				if !game.flags.hasGameStarted {
					initGameStart()
				}
			}
		}
		fmt.Println("Has the Game Started?", game.flags.hasGameStarted)

		if game.flags.isGameLive {
			//init round stats
			initRound()
		}

	})

	p.RegisterEventHandler(func(e events.RoundEnd) {
		fmt.Println("Round End", e.Winner, "won")

		validWinner := true
		if e.Winner < 2 {
			validWinner = false
			//and set the integrity flag to false

		} else if e.Winner == 2 {
			game.flags.tMoney = true
		} else {
			//we need to check if the game is over

			test := (*(e.WinnerState)).ClanName()
			fmt.Println(test, (*(e.WinnerState)).ID())
		}

		//we want to actually process the round
		if game.flags.isGameLive && validWinner && game.flags.roundIntegrityStart == p.GameState().TotalRoundsPlayed()+1 {
			game.rounds[0].winnerENUM = int(e.Winner)
			processRoundOnWinCon((*(e.WinnerState)).ID())

			//check last round
			roundWinnerScore := game.teams[(*(e.WinnerState)).ID()].score + 1
			roundLoserScore := game.teams[(*(e.LoserState)).ID()].score
			fmt.Println("winner Rounds", roundWinnerScore)
			fmt.Println("loser Rounds", roundLoserScore)

			if game.roundsToWin == 16 {
				//check for normal win
				if roundWinnerScore == 16 && roundLoserScore < 15 {
					//normal win
					processRoundFinal(true)
				} else if roundWinnerScore > 15 { //check for OT win
					overtime := ((roundWinnerScore+roundLoserScore)-30-1)/6 + 1
					//OT win
					if (roundWinnerScore-15-1)/3 == overtime {
						processRoundFinal(true)
					}
				}
			} else if game.roundsToWin == 9 {
				//check for normal win
				if roundWinnerScore == 9 && roundLoserScore < 8 {
					//normal win
					processRoundFinal(true)
				} else if roundWinnerScore == 8 && roundLoserScore == 8 { //check for tie
					//tie
					processRoundFinal(true)
				}
			}
		}

		//check last round
		//or check overtime win

	})

	//round end official doesnt fire on the last round
	p.RegisterEventHandler(func(e events.RoundEndOfficial) {
		fmt.Printf("Round End Official\n")

		if game.flags.isGameLive && game.flags.roundIntegrityEnd == p.GameState().TotalRoundsPlayed() {
			processRoundFinal(false)
		}
	})

	// Register handler on kill events
	p.RegisterEventHandler(func(e events.Kill) {
		if game.flags.isGameLive && ((int(game.rounds[0].roundNum) == p.GameState().TotalRoundsPlayed()+1) ||
			(int(game.rounds[0].roundNum) == p.GameState().TotalRoundsPlayed() && game.flags.postWinCon)) {
			pS := game.rounds[0].playerStats
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
					pS[e.Victim.SteamID64].deathPlacement = float64(game.rounds[0].initTerroristCount - game.flags.tAlive)
				} else if e.Victim.Team == 3 {
					game.flags.ctAlive -= 1
					pS[e.Victim.SteamID64].deathPlacement = float64(game.rounds[0].initCTerroristCount - game.flags.ctAlive)
				} else {
					//else log an error
				}

				//check clutch start
				if !game.flags.postWinCon {
					if game.flags.tAlive == 1 && game.flags.tClutchVal == 0 {
						game.flags.tClutchVal = game.flags.ctAlive
						membersT := p.GameState().TeamTerrorists().Members()
						for _, terrorist := range membersT {
							if terrorist.IsAlive() {
								game.flags.tClutchSteam = terrorist.SteamID64
							}
						}
					}
					if game.flags.ctAlive == 1 && game.flags.ctClutchVal == 0 {
						game.flags.ctClutchVal = game.flags.tAlive
						membersCT := p.GameState().TeamCounterTerrorists().Members()
						for _, counterTerrorist := range membersCT {
							if counterTerrorist.IsAlive() {
								game.flags.ctClutchSteam = counterTerrorist.SteamID64
							}
						}
					}
				}

				pS[e.Victim.SteamID64].ticksAlive = tick - game.rounds[0].startingTick
				for deadGuySteam, deadTick := range (*game.rounds[0]).playerStats[e.Victim.SteamID64].tradeList {
					if tick-deadTick < 4*game.tickRate {
						pS[deadGuySteam].traded = 1
						pS[deadGuySteam].kastRounds = 1
					}
				}
			}

			//assist logic
			if assisterExists && victimExists && e.Assister.TeamState.ID() != e.Victim.TeamState.ID() {
				//this logic needs to be replaced
				pS[e.Assister.SteamID64].assists += 1
				pS[e.Assister.SteamID64].kastRounds = 1
				assisted = true
				if e.AssistedFlash {
					pS[e.Assister.SteamID64].fAss += 1
					flashAssisted = true
				}

			}

			//kill logic (trades here)
			if killerExists && victimExists && e.Killer.TeamState.ID() != e.Victim.TeamState.ID() {
				pS[e.Killer.SteamID64].kills += 1
				pS[e.Killer.SteamID64].kastRounds = 1
				pS[e.Killer.SteamID64].tradeList[e.Victim.SteamID64] = tick
				if e.Weapon.Type == 309 {
					pS[e.Killer.SteamID64].awpKills += 1
				}
				if e.IsHeadshot {
					pS[e.Killer.SteamID64].hs += 1
				}
				for _, deadTick := range (*game.rounds[0]).playerStats[e.Victim.SteamID64].tradeList {
					if tick-deadTick < 4*game.tickRate {
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
					if game.rounds[0].winnerENUM == 2 { //Ts win
						if killerTeam == 2 { //chase
							killValue = 0.8
						}
						if killerTeam == 3 { //exit
							killValue = 0.6
						}
					} else if game.rounds[0].winnerENUM == 3 { //CTs win
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
					pS[e.Killer.SteamID64].trades += 1
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
		fmt.Printf("%s <%v%s%s> %s at %d\n", e.Killer, e.Weapon, hs, wallBang, e.Victim, p.GameState().IngameTick())
	})

	p.RegisterEventHandler(func(e events.PlayerHurt) {
		//fmt.Printf("Player Hurt\n")
		if game.flags.isGameLive {
			equipment := e.Weapon.Type
			if e.Player != nil && e.Attacker != nil && e.Player.Team != e.Attacker.Team {
				game.rounds[0].playerStats[e.Attacker.SteamID64].damage += e.HealthDamageTaken
				if equipment >= 500 && equipment <= 506 {
					game.rounds[0].playerStats[e.Attacker.SteamID64].utilDmg += e.HealthDamageTaken
					if equipment == 506 {
						game.rounds[0].playerStats[e.Attacker.SteamID64].nadeDmg += e.HealthDamageTaken
					}
					if equipment == 502 || equipment == 503 {
						game.rounds[0].playerStats[e.Attacker.SteamID64].infernoDmg += e.HealthDamageTaken
					}
				}
			}
		}
	})

	p.RegisterEventHandler(func(e events.PlayerFlashed) {
		//fmt.Printf("Player Flashed\n")
		tick := float64(p.GameState().IngameTick())
		blindTicks := e.FlashDuration().Seconds() * 128.0
		if game.flags.isGameLive && e.Player != nil && e.Attacker != nil {
			victim := e.Player
			flasher := e.Attacker
			if flasher.Team != victim.Team && blindTicks > 128.0 && victim.IsAlive() && (float64(victim.FlashDuration) < (blindTicks/128.0 + 1)) {
				game.rounds[0].playerStats[flasher.SteamID64].ef += 1
				game.rounds[0].playerStats[flasher.SteamID64].enemyFlashTime += blindTicks
				if tick+blindTicks > game.rounds[0].playerStats[victim.SteamID64].mostRecentFlashVal {
					game.rounds[0].playerStats[victim.SteamID64].mostRecentFlashVal = tick + blindTicks
					game.rounds[0].playerStats[victim.SteamID64].mostRecentFlasher = flasher.SteamID64
				}

			}
			if flasher.Name != "" {
				debugMsg := fmt.Sprintf("%s flashed %s for %.2f at %d. He was %f blind.\n", flasher, victim, blindTicks/128, int(tick), victim.FlashDuration)
				debugFile.WriteString(debugMsg)
				debugFile.Sync()
			}

		}
	})

	p.RegisterEventHandler(func(e events.PlayerJump) {
		//fmt.Printf("Player Jumped\n")
	})

	p.RegisterEventHandler(func(e events.BombPlanted) {
		fmt.Printf("Bomb Planted\n")
		if game.flags.isGameLive && game.flags.postWinCon == false {
			game.flags.prePlant = false
			game.flags.postPlant = true
			game.flags.postWinCon = false
			game.flags.tMoney = true
			game.rounds[0].planter = e.BombEvent.Player.SteamID64
		}
	})

	p.RegisterEventHandler(func(e events.BombDefused) {
		fmt.Println("Bomb Defused by", e.BombEvent.Player.Name)
		if game.flags.isGameLive {
			game.flags.prePlant = false
			game.flags.postPlant = false
			game.flags.postWinCon = true
			game.rounds[0].playerStats[e.BombEvent.Player.SteamID64].impactPoints += 0.5
		}
	})

	p.RegisterEventHandler(func(e events.BombExplode) {
		fmt.Printf("Bomb Exploded\n")
		if game.flags.isGameLive {
			game.flags.prePlant = false
			game.flags.postPlant = false
			game.flags.postWinCon = true
			game.rounds[0].playerStats[game.rounds[0].planter].impactPoints += 0.5
		}
	})

	p.RegisterEventHandler(func(e events.GrenadeProjectileThrow) {
		//fmt.Println("Grenade Thrown", e.Projectile.WeaponInstance.Type)
	})

	p.RegisterEventHandler(func(e events.PlayerDisconnected) {
		//fmt.Println("Player DC", e.Player)

		//update alive players
		if game.flags.isGameLive {
			game.flags.tAlive = 0
			game.flags.ctAlive = 0

			membersT := p.GameState().TeamTerrorists().Members()
			for _, terrorist := range membersT {
				if terrorist.IsAlive() {
					game.flags.tAlive += 1
				}
			}
			membersCT := p.GameState().TeamCounterTerrorists().Members()
			for _, counterTerrorist := range membersCT {
				if counterTerrorist.IsAlive() {
					game.flags.ctAlive += 1
				}
			}
		}

	})

	if printChatLog {
		p.RegisterEventHandler(func(e events.ChatMessage) {
			chatMsg := fmt.Sprintf("%s: \"%s\" at %d\n", e.Sender, e.Text, p.GameState().IngameTick())
			chatFile.WriteString(chatMsg)
			chatFile.Sync()
			//check(err)
		})
	}

	// Parse to end
	err = p.ParseToEnd()
	if err != nil {
		panic(err)
	}

	//----END OF MATCH PROCESSING----
	//we want to iterate through rounds backwards to make sure their are no repeats

	game.totalPlayerStats = make(map[uint64]*playerStats)

	validRoundsMap := make(map[int8]bool)
	for i := len(game.rounds) - 1; i >= 0; i-- {
		_, validRoundExists := validRoundsMap[game.rounds[i].roundNum]
		if game.rounds[i].integrityCheck && !validRoundExists {
			//this i-th round is good to add

			validRoundsMap[game.rounds[i].roundNum] = true
			//add to round master stats
			fmt.Println(game.rounds[i].roundNum)
			for steam, player := range (*game.rounds[i]).playerStats {
				if game.totalPlayerStats[steam] == nil {
					game.totalPlayerStats[steam] = &playerStats{name: player.name, steamID: player.steamID}
				}
				game.totalPlayerStats[steam].rounds += 1
				game.totalPlayerStats[steam].kills += player.kills
				game.totalPlayerStats[steam].assists += player.assists
				game.totalPlayerStats[steam].deaths += player.deaths
				game.totalPlayerStats[steam].damage += player.damage
				game.totalPlayerStats[steam].ticksAlive += player.ticksAlive
				game.totalPlayerStats[steam].deathPlacement += player.deathPlacement
				game.totalPlayerStats[steam].trades += player.trades
				game.totalPlayerStats[steam].traded += player.traded
				game.totalPlayerStats[steam].ok += player.ok
				game.totalPlayerStats[steam].ol += player.ol
				game.totalPlayerStats[steam].killPoints += player.killPoints
				game.totalPlayerStats[steam].cl_1 += player.cl_1
				game.totalPlayerStats[steam].cl_2 += player.cl_2
				game.totalPlayerStats[steam].cl_3 += player.cl_3
				game.totalPlayerStats[steam].cl_4 += player.cl_4
				game.totalPlayerStats[steam].cl_5 += player.cl_5
				game.totalPlayerStats[steam]._2k += player._2k
				game.totalPlayerStats[steam]._3k += player._3k
				game.totalPlayerStats[steam]._4k += player._4k
				game.totalPlayerStats[steam]._5k += player._5k
				game.totalPlayerStats[steam].nadeDmg += player.nadeDmg
				game.totalPlayerStats[steam].infernoDmg += player.infernoDmg
				game.totalPlayerStats[steam].utilDmg += player.utilDmg
				game.totalPlayerStats[steam].ef += player.ef
				game.totalPlayerStats[steam].fAss += player.fAss
				game.totalPlayerStats[steam].enemyFlashTime += player.enemyFlashTime
				game.totalPlayerStats[steam].hs += player.hs
				game.totalPlayerStats[steam].kastRounds += player.kastRounds
				game.totalPlayerStats[steam].saves += player.saves
				game.totalPlayerStats[steam].entries += player.entries
				game.totalPlayerStats[steam].impactPoints += player.impactPoints
				game.totalPlayerStats[steam].winPoints += player.winPoints

				//game.rounds[0].teamStats[].points +=]

			}
		}
	}

	//check our shit
	for _, player := range game.totalPlayerStats {
		player.atd = player.ticksAlive / player.rounds / game.tickRate
		player.deathPlacement = player.deathPlacement / float64(player.rounds)
		player.kast = player.kastRounds / float64(player.rounds)
		player.killPointAvg = player.killPoints / float64(player.kills)
		fmt.Printf("%+v\n", player)
	}

	fmt.Println("Demo is complete!")
	//cleanup()

}
