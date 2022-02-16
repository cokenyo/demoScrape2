package main

import (
	"fmt"
	"strings"

	dem "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
	common "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/common"
)

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

func calculateTeamEquipmentValue(game *game, team *common.TeamState) int {
	equipment := 0
	for _, teamMember := range team.Members() {
		if teamMember.IsAlive() && game.potentialRound.playerStats[teamMember.SteamID64].health > 0 {
			equipment += teamMember.EquipmentValueCurrent()
		}
	}
	return equipment
}

//works for grenades, needs to be modified for other types
func calculateTeamEquipmentNum(game *game, team *common.TeamState, equipmentENUM int) int {
	equipment := 0
	for _, teamMember := range team.Members() {
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

func closestCTDisttoBomb(game *game, team *common.TeamState, bomb *common.Bomb) int {
	var distance int = 999999
	for _, teamMember := range team.Members() {
		if teamMember.IsAlive() && game.potentialRound.playerStats[teamMember.SteamID64].health > 0 {
			if bomb.Position().Distance(teamMember.Position()) < float64(distance) {
				distance = int(bomb.Position().Distance(teamMember.Position()))
			}
		}
	}
	return distance
}

func numOfKits(game *game, team *common.TeamState) int {
	kits := 0
	for _, teamMember := range team.Members() {
		if teamMember.IsAlive() && game.potentialRound.playerStats[teamMember.SteamID64].health > 0 {
			if teamMember.HasDefuseKit() {
				kits += 1
			}
		}
	}
	return kits
}

func playersWithArmor(game *game, team *common.TeamState) int {
	armor := 0
	for _, teamMember := range team.Members() {
		if teamMember.IsAlive() && game.potentialRound.playerStats[teamMember.SteamID64].health > 0 {
			if teamMember.Armor() > 0 {
				armor += 1
			}
		}
	}
	return armor
}
