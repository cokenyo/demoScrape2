package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/machinebox/graphql"
)

type Dictionary map[string]interface{}

func authenticate() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	graphqlEndpoint := os.Getenv("GRAPHQL_ENDPOINT")
	graphqlDiscord := os.Getenv("GRAPHQL_DISCORDID")
	graphqlPass := os.Getenv("GRAPHQL_PASSWORD")

	client := graphql.NewClient(graphqlEndpoint)

	req := graphql.NewRequest(`
		mutation Authenticate($discordId: String!, $password: String!) {
			tokenAuth(discordId: $discordId, password: $password) {
				token
			}
		}
	`)

	req.Var("discordId", graphqlDiscord)
	req.Var("password", graphqlPass)


	ctx := context.Background()

	type ResponseStruct struct {
		TokenAuth struct {
			Token string
		}
	}

	var respData ResponseStruct
	if err := client.Run(ctx, req, &respData); err != nil {
		fmt.Println(err)
	}

	fmt.Println(respData)


	return respData.TokenAuth.Token

	//return m.TokenAuth.Token
}

func verifyOriginalMatch(token string, coreID string) bool {
	fmt.Println("VERIFYING ORIGINAL MATCH")

	client := graphql.NewClient(os.Getenv("GRAPHQL_ENDPOINT"))

	req := graphql.NewRequest(`
		query GetMatch($coreId: String!) {
			match(matchId: $coreId) {
				id
				stats {
					id
				}
			}
		}
	`)

	req.Var("coreId", coreID)
	req.Header.Set("Authorization", "JWT "+token)

	type ResponseStruct struct {
		Match struct {
			Id string
			Stats struct {
				Id string
			}
		}
	}

	var respData ResponseStruct
	if err := client.Run(context.Background(), req, &respData); err != nil {
		fmt.Println(err)
	}

	return respData.Match.Id != "" && respData.Match.Stats.Id == ""

	// fmt.Println(q.Match)

	// // I think this will explode!
	// return q.Match.Id != nil && q.Match.Stats.Id == nil
}

func addMatchStats(token string, game *game) {
	fmt.Println("ADDING MATCH")

	client := graphql.NewClient(os.Getenv("GRAPHQL_ENDPOINT"))

	req := graphql.NewRequest(`
		mutation AddMatchStats($matchId: String!, $winnerTeamName: String!, $score: String!, $matchStatsInput: MatchStatsInput!) {
			addMatchStats(matchId: $matchId, winnerTeamName: $winnerTeamName, score: $score, matchStatsInput: $matchStatsInput) {
				id
			}
		}
	`)

	matchStatsInput := Dictionary{
		"playerStats": []Dictionary{
		},
		"rounds": []Dictionary{
		},
		"teamStats": []Dictionary{
		},
	}

	for _, player := range game.totalPlayerStats {
		playerStat := Dictionary{
			"playerSteamId": strconv.FormatUint(player.steamID, 10),
			"side": "t",
			"adp": player.tADP,
			"adr": player.tADR,
			"assists": player.assists,
			"atd": player.atd,
			"awpK": player.awpKills,
			"damageDealt": player.tDamage,
			"damageTaken": player.damageTaken,
			"deaths": player.tDeaths,
			"ef": player.ef,
			"eft": player.enemyFlashTime,
			"fAss": player.fAss,
			"fDeaths": player.tOL,
			"fireDamage": player.infernoDmg,
			"fires": player.firesThrown,
			"fiveK": player._5k,
			"fourK": player._4k,
			"threeK": player._3k,
			"twoK": player._2k,
			"fKills": player.tOK,
			"flashes": player.flashThrown,
			"hs": player.hs,
			"impact": player.impactRating,
			"iwr": player.iiwr,
			"jumps": 0,
			"kast": player.kast,
			"kills": player.kills,
			"kpa": player.killPointAvg,
			"lurks": player.lurkRounds,
			"mip": player.mip,
			"nadeDamage": player.nadeDmg,
			"nades": player.nadesThrown,
			"oneVFive": player.cl_5,
			"oneVFour": player.cl_4,
			"oneVThree": player.cl_3,
			"oneVTwo": player.cl_2,
			"oneVOne": player.cl_1,
			"ra": player.RA,
			"rating": player.rating,
			"rf": player.RF,
			"rounds": player.tRounds,
			"rwk": player.rwk,
			"saves": player.saves,
			"smokes": player.smokeThrown,
			"suppR": player.suppRounds,
			"suppX": player.suppDamage,
			"traded": player.traded,
			"trades": player.trades,
			"ud": player.utilDmg,
			"util": player.utilThrown,
			"wlp": player.wlp,
		}

		matchStatsInput["playerStats"] = append(matchStatsInput["playerStats"].([]Dictionary), playerStat)
	}

	for _, round := range game.rounds {
		roundStat := Dictionary{
			"ctPlayers": round.initCTerroristCount,
			"tPlayers": round.initTerroristCount,
			"defuserSteamId": strconv.FormatUint(round.defuser, 10),
			"planterSteamId": strconv.FormatUint(round.planter, 10),
			"roundLength": round.endingTick,
			"roundNumber": round.roundNum,
			"roundWinnerTeamName": round.winnerClanName,
			"sideWinner": strconv.Itoa(round.winnerENUM),
		}

		matchStatsInput["rounds"] = append(matchStatsInput["rounds"].([]Dictionary), roundStat)
	}

	for name, team := range game.totalTeamStats {
		teamStat := Dictionary{
			"teamName": name,
			"clutches": team.clutches,
			"deaths": team.deaths,
			"ef": team.ef,
			"fa": team.fass,
			"fourVFives": team._4v5s,
			"pistolsWon": team.pistolsW,
			"rounds": 0,
			"roundsAgainst": 0,
			"roundsWon": 0,
			"saves": 0,
			"side": "t",
			"traded": team.traded,
			"ud": team.ud,
			"util": team.util,
			"wonFourVFives": team._4v5w,
		}

		matchStatsInput["teamStats"] = append(matchStatsInput["teamStats"].([]Dictionary), teamStat)
	}

	req.Var("matchId", game.coreID)
	req.Var("winnerTeamName", game.winnerClanName)
	req.Var("score", "16:14")
	req.Var("matchStatsInput", matchStatsInput)

	req.Header.Set("Authorization", "JWT "+token)

	type ResponseStruct struct {
		AddMatchStats struct {
			Id string
		}
	}

	var respData ResponseStruct
	if err := client.Run(context.Background(), req, &respData); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("We created match stats with ID: " + respData.AddMatchStats.Id)
}
