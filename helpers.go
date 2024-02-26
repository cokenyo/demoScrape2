package main

import (
	"fmt"
	"strconv"
	"strings"

	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	common "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
)

func getTeamMembers(team *common.TeamState, game *game, p dem.Parser) []*common.Player {
	players := team.Members()
	allPlayers := p.GameState().Participants().All()
	// Filter players by the Team from the team state
	teamPlayers := make([]*common.Player, 0)

	// Helper function to find player index in teamPlayers
	findPlayerIndex := func(slice []*common.Player, steamId uint64) int {
		for i, player := range slice {
			if player.SteamID64 == steamId {
				return i
			}
		}
		return -1
	}

	for _, player := range players {
		if player.Team == team.Team() {
			if game.connectedAfterRoundStart[player.SteamID64] {
				continue
			}
			teamPlayers = append(teamPlayers, player)
		}
	}

	// Grab reconnected players and check for duplicates
	for steamId, connected := range game.reconnectedPlayers {
		if !connected {
			continue
		}
		for _, player := range allPlayers {
			// If the player is in connectedAfterRoundStart, do not return them
			if game.connectedAfterRoundStart[player.SteamID64] {
				continue
			}

			if player.SteamID64 == steamId && player.Team == team.Team() {
				// Check if player is already in teamPlayers
				idx := findPlayerIndex(teamPlayers, steamId)
				if idx != -1 {
					// Remove the existing record
					teamPlayers = append(teamPlayers[:idx], teamPlayers[idx+1:]...)
				}
				// Append the new record
				teamPlayers = append(teamPlayers, player)
			}
		}
	}

	return teamPlayers
}

func isDuringExpectedRound(game *game, p dem.Parser) bool {
	isPreWinCon := (int(game.potentialRound.roundNum) == p.GameState().TotalRoundsPlayed()+1)
	isAfterWinCon := (int(game.potentialRound.roundNum) == p.GameState().TotalRoundsPlayed() && game.flags.postWinCon)
	return (isPreWinCon || isAfterWinCon)
}

func validateTeamName(game *game, teamName string, teamNum common.Team) string {
	if teamName != "" {
		name := ""
		if strings.HasPrefix(teamName, "[") {
			if len(teamName) == 31 {
				//name here will be truncated
				name = strings.Split(teamName, "] ")[1]
				for _, team := range game.teams {
					if strings.Contains(team.name, name) {
						return team.name
					}
				}
				fmt.Print("OH NOEY")
				return name
			} else {
				name = strings.Split(teamName, "] ")[1]
				return name
			}
		} else {
			return teamName
		}
	} else {
		//this demo no have team names, so we are big fucked
		//we are hardcoding during what rounds each team will have what side
		round := game.potentialRound.roundNum
		swap := false
		if round >= 16 && round <= 33 {
			swap = true
		} else if round >= 34 {
			//we are now in OT hell :)
			if (round-34)/6%2 != 0 {
				swap = true
			}
		}
		if !swap {
			if teamNum == 2 {
				return "StartedT"
			} else if teamNum == 3 {
				return "StartedCT"
			}
		} else {
			if teamNum == 2 {
				return "StartedCT"
			} else if teamNum == 3 {
				return "StartedT"
			}
			return "SPECs"
		}
		return "SPECs"
	}
}

func calculateTeamEquipmentValue(game *game, team *common.TeamState, p dem.Parser) int {
	equipment := 0
	for _, teamMember := range getTeamMembers(team, game, p) {
		if teamMember.IsAlive() && game.potentialRound.playerStats[teamMember.SteamID64].health > 0 {
			equipment += teamMember.EquipmentValueCurrent()
		}
	}
	return equipment
}

// works for grenades, needs to be modified for other types
func calculateTeamEquipmentNum(game *game, team *common.TeamState, equipmentENUM int, p dem.Parser) int {
	equipment := 0
	for _, teamMember := range getTeamMembers(team, game, p) {
		if teamMember.IsAlive() && game.potentialRound.playerStats[teamMember.SteamID64].health > 0 {
			//fmt.Println(teamMember.Inventory)
			//fmt.Println(teamMember.Weapons())
			//fmt.Println(teamMember.AmmoLeft)
			//gren := teamMember.Inventory[equipmentENUM]
			equipment += teamMember.AmmoLeft[equipmentENUM]
		}
	}
	return equipment
}

func closestCTDisttoBomb(game *game, team *common.TeamState, bomb *common.Bomb, p dem.Parser) int {
	var distance int = 999999
	for _, teamMember := range getTeamMembers(team, game, p) {
		if teamMember.IsAlive() && game.potentialRound.playerStats[teamMember.SteamID64].health > 0 {
			if bomb.Position().Distance(teamMember.Position()) < float64(distance) {
				distance = int(bomb.Position().Distance(teamMember.Position()))
			}
		}
	}
	return distance
}

func numOfKits(game *game, team *common.TeamState, p dem.Parser) int {
	kits := 0
	for _, teamMember := range getTeamMembers(team, game, p) {
		if teamMember.IsAlive() && game.potentialRound.playerStats[teamMember.SteamID64].health > 0 {
			if teamMember.HasDefuseKit() {
				kits += 1
			}
		}
	}
	return kits
}

func playersWithArmor(game *game, team *common.TeamState, p dem.Parser) int {
	armor := 0
	for _, teamMember := range getTeamMembers(team, game, p) {
		if teamMember.IsAlive() && game.potentialRound.playerStats[teamMember.SteamID64].health > 0 {
			if teamMember.Armor() > 0 {
				armor += 1
			}
		}
	}
	return armor
}

var roundEndReasons = map[int]string{
	0:  "StillInProgress", //base values
	1:  "TargetBombed",
	2:  "VIPEscaped",
	3:  "VIPKilled",
	4:  "TerroristsEscaped",
	5:  "CTStoppedEscape",
	6:  "TerroristsStopped",
	7:  "BombDefused",
	8:  "CTWin",
	9:  "TerroristsWin",
	10: "Draw",
	11: "HostagesRescued",
	12: "TargetSaved",
	13: "HostagesNotRescued",
	14: "TerroristsNotEscaped",
	15: "VIPNotEscaped",
	16: "GameStart",
	17: "TerroristsSurrender",
	18: "CTSurrender",
	19: "TerroristsPlanted",
	20: "CTsReachedHostage",
}

func getPlayerAPIDict(side string, player *playerStats) Dictionary {

	return Dictionary{
		"playerSteamId": strconv.FormatUint(player.steamID, 10),
		"side":          side,
		"teamName":      player.teamClanName,
		"adp":           player.deathPlacement,
		"adr":           player.adr,
		"assists":       player.assists,
		"atd":           player.atd,
		"awpK":          player.awpKills,
		"damageDealt":   player.damage,
		"damageTaken":   player.damageTaken,
		"deaths":        player.deaths,
		"eac":           player.eac,
		"ef":            player.ef,
		"eft":           player.enemyFlashTime,
		"fAss":          player.fAss,
		"fDeaths":       player.ol,
		"fireDamage":    player.infernoDmg,
		"fires":         player.firesThrown,
		"fiveK":         player._5k,
		"fourK":         player._4k,
		"threeK":        player._3k,
		"twoK":          player._2k,
		"fKills":        player.ok,
		"flashes":       player.flashThrown,
		"hs":            player.hs,
		"impact":        player.impactRating,
		"iwr":           player.iiwr,
		"jumps":         0,
		"kast":          player.kast,
		"kills":         player.kills,
		"kpa":           player.killPointAvg,
		"lurks":         player.lurkRounds,
		"mip":           player.mip,
		"nadeDamage":    player.nadeDmg,
		"nades":         player.nadesThrown,
		"oneVFive":      player.cl_5,
		"oneVFour":      player.cl_4,
		"oneVThree":     player.cl_3,
		"oneVTwo":       player.cl_2,
		"oneVOne":       player.cl_1,
		"ra":            player.RA,
		"rating":        player.rating,
		"rf":            player.RF,
		"rounds":        player.rounds,
		"rwk":           player.rwk,
		"rws":           player.rws,
		"saves":         player.saves,
		"smokes":        player.smokeThrown,
		"suppR":         player.suppRounds,
		"suppX":         player.suppDamage,
		"traded":        player.traded,
		"trades":        player.trades,
		"ud":            player.utilDmg,
		"util":          player.utilThrown,
		"wlp":           player.wlp,
	}
}
