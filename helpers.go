package main

import (
	dem "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
)

func isDuringExpectedRound(game *game, p dem.Parser) bool {
	isPreWinCon := (int(game.potentialRound.roundNum) == p.GameState().TotalRoundsPlayed()+1)
	isAfterWinCon := (int(game.potentialRound.roundNum) == p.GameState().TotalRoundsPlayed() && game.flags.postWinCon)
	return (isPreWinCon || isAfterWinCon)
}

func intInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
