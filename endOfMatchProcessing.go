package main

import (
	"fmt"
	//dem "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
)

func endOfMatchProcessing(game *game) {

	game.totalPlayerStats = make(map[uint64]*playerStats)
	game.totalTeamStats = make(map[string]*teamStats)

	validRoundsMap := make(map[int8]bool)
	for i := len(game.rounds) - 1; i >= 0; i-- {
		_, validRoundExists := validRoundsMap[game.rounds[i].roundNum]
		if game.rounds[i].integrityCheck && !validRoundExists {
			//this i-th round is good to add

			validRoundsMap[game.rounds[i].roundNum] = true
			game.rounds[i].serverNormalizer += game.rounds[i].initTerroristCount + game.rounds[i].initCTerroristCount

			for teamName, team := range game.rounds[i].teamStats {
				if game.totalTeamStats[teamName] == nil {
					game.totalTeamStats[teamName] = &teamStats{}
				}
				game.totalTeamStats[teamName].pistols += team.pistols
				game.totalTeamStats[teamName].pistolsW += team.pistolsW
				game.totalTeamStats[teamName]._4v5s += team._4v5s
				game.totalTeamStats[teamName]._4v5w += team._4v5w
				game.totalTeamStats[teamName]._5v4s += team._5v4s
				game.totalTeamStats[teamName]._5v4w += team._5v4w
				game.totalTeamStats[teamName].saves += team.saves
				game.totalTeamStats[teamName].clutches += team.clutches
				game.totalTeamStats[teamName].ctR += team.ctR
				game.totalTeamStats[teamName].ctRW += team.ctRW
				game.totalTeamStats[teamName].tR += team.tR
				game.totalTeamStats[teamName].tRW += team.tRW
			}

			//add to round master stats
			fmt.Println(game.rounds[i].roundNum)
			for steam, player := range (*game.rounds[i]).playerStats {
				if game.totalPlayerStats[steam] == nil {
					game.totalPlayerStats[steam] = &playerStats{name: player.name, steamID: player.steamID, teamClanName: player.teamClanName}
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
				game.totalPlayerStats[steam].damageTaken += player.damageTaken
				game.totalPlayerStats[steam].suppDamage += player.suppDamage
				game.totalPlayerStats[steam].suppRounds += player.suppRounds
				game.totalPlayerStats[steam].rwk += player.rwk
				game.totalPlayerStats[steam].mip += player.mip

				if player.side == 2 {
					game.totalPlayerStats[steam].winPointsNormalizer += game.rounds[i].initTerroristCount
					game.totalPlayerStats[steam].tImpactPoints += player.impactPoints
					game.totalPlayerStats[steam].tWinPoints += player.winPoints
					game.totalPlayerStats[steam].tOK += player.ok
					game.totalPlayerStats[steam].tOL += player.ol
					game.totalPlayerStats[steam].tKills += player.kills
					game.totalPlayerStats[steam].tDeaths += player.deaths
					game.totalPlayerStats[steam].tKASTRounds += player.kastRounds
					game.totalPlayerStats[steam].tDamage += player.damage
					game.totalPlayerStats[steam].tADP += player.deathPlacement
					//game.totalPlayerStats[steam].tTeamsWinPoints +=
					game.totalPlayerStats[steam].tWinPointsNormalizer += game.rounds[i].initTerroristCount
					game.totalPlayerStats[steam].tRounds += 1
					game.totalPlayerStats[steam].tRF += player.RF
					game.totalPlayerStats[steam].lurkRounds += player.lurkRounds
					if player.lurkRounds != 0 {
						game.totalPlayerStats[steam].wlp += player.winPoints
					}

					game.rounds[i].teamStats[player.teamClanName].tWinPoints += player.winPoints
					game.rounds[i].teamStats[player.teamClanName].tImpactPoints += player.impactPoints
				} else if player.side == 3 {
					game.totalPlayerStats[steam].winPointsNormalizer += game.rounds[i].initCTerroristCount
					game.totalPlayerStats[steam].ctImpactPoints += player.impactPoints
					game.totalPlayerStats[steam].ctWinPoints += player.winPoints
					game.totalPlayerStats[steam].ctOK += player.ok
					game.totalPlayerStats[steam].ctOL += player.ol
					game.totalPlayerStats[steam].ctKills += player.kills
					game.totalPlayerStats[steam].ctDeaths += player.deaths
					game.totalPlayerStats[steam].ctKASTRounds += player.kastRounds
					game.totalPlayerStats[steam].ctDamage += player.damage
					game.totalPlayerStats[steam].ctADP += player.deathPlacement
					//game.totalPlayerStats[steam].tTeamsWinPoints +=
					game.totalPlayerStats[steam].ctWinPointsNormalizer += game.rounds[i].initCTerroristCount
					game.totalPlayerStats[steam].ctRounds += 1
					game.totalPlayerStats[steam].ctAWP += player.ctAWP

					game.rounds[i].teamStats[player.teamClanName].ctWinPoints += player.winPoints
					game.rounds[i].teamStats[player.teamClanName].ctImpactPoints += player.impactPoints
				}

				game.rounds[i].teamStats[player.teamClanName].winPoints += player.winPoints
				game.rounds[i].teamStats[player.teamClanName].impactPoints += player.impactPoints

			}
			for steam, player := range (*game.rounds[i]).playerStats {
				game.totalPlayerStats[steam].teamsWinPoints += game.rounds[i].teamStats[player.teamClanName].winPoints
				game.totalPlayerStats[steam].tTeamsWinPoints += game.rounds[i].teamStats[player.teamClanName].tWinPoints
				game.totalPlayerStats[steam].ctTeamsWinPoints += game.rounds[i].teamStats[player.teamClanName].ctWinPoints
			}
		}
	}

	for _, player := range game.totalPlayerStats {
		game.totalTeamStats[player.teamClanName].util += player.smokeThrown + player.flashThrown + player.nadesThrown + player.firesThrown
		game.totalTeamStats[player.teamClanName].ud += player.utilDmg
		game.totalTeamStats[player.teamClanName].ef += player.ef
		game.totalTeamStats[player.teamClanName].fass += player.fAss
		game.totalTeamStats[player.teamClanName].traded += player.traded
		game.totalTeamStats[player.teamClanName].deaths += int(player.deaths)
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

	tImpactRoundAvg := 0.0
	tKillRoundAvg := 0.0
	tDeathRoundAvg := 0.0
	tKastRoundAvg := 0.0
	tAdrAvg := 0.0
	tRoundNormalizer := 0

	ctImpactRoundAvg := 0.0
	ctKillRoundAvg := 0.0
	ctDeathRoundAvg := 0.0
	ctKastRoundAvg := 0.0
	ctAdrAvg := 0.0
	ctRoundNormalizer := 0

	//check our shit
	for _, player := range game.totalPlayerStats {

		player.atd = player.ticksAlive / player.rounds / game.tickRate
		player.deathPlacement = player.deathPlacement / float64(player.deaths)
		player.kast = player.kastRounds / float64(player.rounds)
		player.killPointAvg = player.killPoints / float64(player.kills)
		player.iiwr = player.winPoints / player.impactPoints
		player.adr = float64(player.damage) / float64(player.rounds)
		player.drDiff = player.adr - (float64(player.damageTaken) / float64(player.rounds))
		player.tr = float64(player.traded) / float64(player.deaths)
		player.kR = float64(player.kills) / float64(player.rounds)
		player.utilThrown = player.smokeThrown + player.flashThrown + player.nadesThrown + player.firesThrown
		player.ctADR = float64(player.ctDamage) / float64(player.ctRounds)
		player.tADR = float64(player.tDamage) / float64(player.tRounds)
		player.tKAST = player.tKASTRounds / float64(player.tRounds)
		player.ctKAST = player.ctKASTRounds / float64(player.ctRounds)
		player.tADP = player.tADP / float64(player.tDeaths)
		player.ctADP = player.ctADP / float64(player.ctDeaths)

		roundNormalizer += player.rounds
		impactRoundAvg += player.impactPoints
		killRoundAvg += float64(player.kills)
		deathRoundAvg += float64(player.deaths)
		kastRoundAvg += player.kastRounds
		adrAvg += float64(player.damage)

		tImpactRoundAvg += player.tImpactPoints
		tKillRoundAvg += float64(player.tKills)
		tDeathRoundAvg += float64(player.ctDeaths)
		tKastRoundAvg += player.tKASTRounds
		tAdrAvg += float64(player.tDamage)
		tRoundNormalizer += player.tRounds

		ctImpactRoundAvg += player.ctImpactPoints
		ctKillRoundAvg += float64(player.ctKills)
		ctDeathRoundAvg += float64(player.tDeaths)
		ctKastRoundAvg += player.ctKASTRounds
		ctAdrAvg += float64(player.ctDamage)
		ctRoundNormalizer += player.ctRounds

	}

	impactRoundAvg /= float64(roundNormalizer)
	killRoundAvg /= float64(roundNormalizer)
	deathRoundAvg /= float64(roundNormalizer)
	kastRoundAvg /= float64(roundNormalizer)
	adrAvg /= float64(roundNormalizer)

	tImpactRoundAvg /= float64(tRoundNormalizer)
	tKillRoundAvg /= float64(tRoundNormalizer)
	tDeathRoundAvg /= float64(tRoundNormalizer)
	tKastRoundAvg /= float64(tRoundNormalizer)
	tAdrAvg /= float64(tRoundNormalizer)

	ctImpactRoundAvg /= float64(ctRoundNormalizer)
	ctKillRoundAvg /= float64(ctRoundNormalizer)
	ctDeathRoundAvg /= float64(ctRoundNormalizer)
	ctKastRoundAvg /= float64(ctRoundNormalizer)
	ctAdrAvg /= float64(ctRoundNormalizer)

	for _, player := range game.totalPlayerStats {
		openingFactor := (float64(player.ok-player.ol) / 13.0) + 1 //move from 13 to (rounds / 5)
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

		//ctRating
		openingFactor = (float64(player.ctOK-player.ctOL) / 13.0) + 1
		playerIPR = player.ctImpactPoints / float64(player.ctRounds)
		playerWPR = player.ctWinPoints / float64(player.ctRounds)

		ctImpactRating := 0.0
		if player.ctTeamsWinPoints != 0 {
			ctImpactRating = (0.1 * float64(openingFactor)) + (0.6 * (playerIPR / ctImpactRoundAvg)) + (0.3 * (playerWPR / (player.ctTeamsWinPoints / float64(player.ctWinPointsNormalizer))))
		} else {
			fmt.Println("UH 16-0?")
			ctImpactRating = (0.1 * float64(openingFactor)) + (0.6 * (playerIPR / ctImpactRoundAvg))
		}
		playerDR = float64(player.ctDeaths) / float64(player.ctRounds)
		player.ctRating = (0.3 * ctImpactRating) + (0.35 * ((float64(player.ctKills) / float64(player.ctRounds)) / ctKillRoundAvg)) + (0.07 * (ctDeathRoundAvg / playerDR)) + (0.08 * (player.ctKAST / ctKastRoundAvg)) + (0.2 * (player.ctADR / ctAdrAvg))

		//tRating
		openingFactor = (float64(player.tOK-player.tOL) / 13.0) + 1
		playerIPR = player.tImpactPoints / float64(player.tRounds)
		playerWPR = player.tWinPoints / float64(player.tRounds)

		tImpactRating := 0.0
		if player.tTeamsWinPoints != 0 {
			tImpactRating = (0.1 * float64(openingFactor)) + (0.6 * (playerIPR / tImpactRoundAvg)) + (0.3 * (playerWPR / (player.tTeamsWinPoints / float64(player.tWinPointsNormalizer))))
		} else {
			fmt.Println("UH 16-0?")
			tImpactRating = (0.1 * float64(openingFactor)) + (0.6 * (playerIPR / tImpactRoundAvg))
		}
		playerDR = float64(player.tDeaths) / float64(player.tRounds)
		player.tRating = (0.3 * tImpactRating) + (0.35 * ((float64(player.tKills) / float64(player.tRounds)) / tKillRoundAvg)) + (0.07 * (tDeathRoundAvg / playerDR)) + (0.08 * (player.tKAST / tKastRoundAvg)) + (0.2 * (player.tADR / tAdrAvg))

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

	beginOutput(game)
}
