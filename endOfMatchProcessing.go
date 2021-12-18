package main

import (
	"fmt"
	//dem "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
)

func endOfMatchProcessing(game *game) {

	game.totalPlayerStats = make(map[uint64]*playerStats)

	validRoundsMap := make(map[int8]bool)
	for i := len(game.rounds) - 1; i >= 0; i-- {
		_, validRoundExists := validRoundsMap[game.rounds[i].roundNum]
		if game.rounds[i].integrityCheck && !validRoundExists {
			//this i-th round is good to add

			validRoundsMap[game.rounds[i].roundNum] = true
			game.rounds[i].serverNormalizer += game.rounds[i].initTerroristCount + game.rounds[i].initCTerroristCount

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
				game.totalPlayerStats[steam].awpKills += player.awpKills
				game.totalPlayerStats[steam].RF += player.RF
				game.totalPlayerStats[steam].RA += player.RA
				game.totalPlayerStats[steam].nadesThrown += player.nadesThrown
				game.totalPlayerStats[steam].smokeThrown += player.smokeThrown
				game.totalPlayerStats[steam].flashThrown += player.flashThrown
				game.totalPlayerStats[steam].firesThrown += player.firesThrown

				if player.side == 2 {
					game.totalPlayerStats[steam].winPointsNormalizer += game.rounds[i].initTerroristCount
				} else if player.side == 3 {
					game.totalPlayerStats[steam].winPointsNormalizer += game.rounds[i].initCTerroristCount
				}

				game.rounds[i].teamStats[player.teamID].winPoints += player.winPoints
				game.rounds[i].teamStats[player.teamID].impactPoints += player.impactPoints

			}
			for steam, player := range (*game.rounds[i]).playerStats {
				game.totalPlayerStats[steam].teamsWinPoints += game.rounds[i].teamStats[player.teamID].winPoints

			}
		}
	}

	calculateDerivedFields(game)

}

func calculateDerivedFields(game *game) {

	impactRoundAvg := 0.0
	killRoundAvg := 0.0
	deathRoundAvg := 0.0
	kastRoundAvg := 0.0
	adrAvg := 0.0
	roundNormalizer := 0

	//check our shit
	for _, player := range game.totalPlayerStats {

		player.atd = player.ticksAlive / player.rounds / game.tickRate
		player.deathPlacement = player.deathPlacement / float64(player.rounds)
		player.kast = player.kastRounds / float64(player.rounds)
		player.killPointAvg = player.killPoints / float64(player.kills)
		player.iiwr = player.winPoints / player.impactPoints
		player.adr = float64(player.damage) / float64(player.rounds)
		player.tr = float64(player.traded) / float64(player.deaths)
		player.kR = float64(player.kills) / float64(player.rounds)
		player.utilThrown = player.smokeThrown + player.flashThrown + player.nadesThrown + player.firesThrown

		roundNormalizer += player.rounds
		impactRoundAvg += player.impactPoints
		killRoundAvg += float64(player.kills)
		deathRoundAvg += float64(player.deaths)
		kastRoundAvg += player.kastRounds
		adrAvg += float64(player.damage)

	}

	impactRoundAvg /= float64(roundNormalizer)
	killRoundAvg /= float64(roundNormalizer)
	deathRoundAvg /= float64(roundNormalizer)
	kastRoundAvg /= float64(roundNormalizer)
	adrAvg /= float64(roundNormalizer)
	fmt.Println(impactRoundAvg)

	for _, player := range game.totalPlayerStats {
		openingFactor := (float64(player.ok-player.ol) / 13.0) + 1
		playerIPR := player.impactPoints / float64(player.rounds)
		playerWPR := player.winPoints / float64(player.rounds)

		if player.teamsWinPoints != 0 {
			player.impactRating = (0.1 * float64(openingFactor)) + (0.6 * (playerIPR / impactRoundAvg)) + (0.3 * (playerWPR / (player.teamsWinPoints / float64(player.winPointsNormalizer))))
		} else {
			fmt.Println("UH 16-0?")
			player.impactRating = (0.1 * float64(openingFactor)) + (0.6 * (playerIPR / impactRoundAvg))
		}
		playerDR := float64(player.deaths) / float64(player.rounds)
		player.rating = (0.3 * player.impactRating) + (0.35 * (player.kR / killRoundAvg)) + (0.07 * (deathRoundAvg / playerDR)) + (0.08 * (player.kast / kastRoundAvg)) + (0.2 * (player.adr / adrAvg))

		fmt.Println("openingFactor", (0.1 * float64(openingFactor)))
		fmt.Println("playerIPR", (0.6 * (playerIPR / impactRoundAvg)))
		fmt.Println("playerWPR", (0.3 * (playerWPR / (player.teamsWinPoints / float64(player.winPointsNormalizer)))))
		fmt.Println("player.teamsWinPoints", player.teamsWinPoints)
		fmt.Println("player.winPointsNormalizer", player.winPointsNormalizer)

		fmt.Printf("%+v\n\n", player)
	}
	fmt.Println("impactRoundAvg", impactRoundAvg)
	fmt.Println("killRoundAvg", killRoundAvg)
	fmt.Println("deathRoundAvg", deathRoundAvg)
	fmt.Println("kastRoundAvg", kastRoundAvg)
	fmt.Println("adrAvg", adrAvg)
}
