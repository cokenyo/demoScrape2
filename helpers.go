package main

import (
	"fmt"
	"strings"

	dem "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
)

func isDuringExpectedRound(game *game, p dem.Parser) bool {
	isPreWinCon := (int(game.potentialRound.roundNum) == p.GameState().TotalRoundsPlayed()+1)
	isAfterWinCon := (int(game.potentialRound.roundNum) == p.GameState().TotalRoundsPlayed() && game.flags.postWinCon)
	return (isPreWinCon || isAfterWinCon)
}

func validateTeamName(game *game, teamName string) string {
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
}
