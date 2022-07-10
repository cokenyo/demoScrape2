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

	//fmt.Println(respData)

	return respData.TokenAuth.Token

	//return m.TokenAuth.Token
}

func verifyOriginalMatch(token string, coreID string, mapNumber int) bool {
	fmt.Println("VERIFYING ORIGINAL MATCH")

	client := graphql.NewClient(os.Getenv("GRAPHQL_ENDPOINT"))

	req := graphql.NewRequest(`
		query GetMatch($coreId: String!) {
			match(matchId: $coreId) {
				id
				stats {
					id
					mapNumber
				}
			}
		}
	`)

	req.Var("coreId", coreID)
	req.Header.Set("Authorization", "JWT "+token)

	type ResponseStruct struct {
		Match struct {
			Id    string
			Stats []struct {
				Id        string
				MapNumber int
			}
		}
	}

	var respData ResponseStruct
	if err := client.Run(context.Background(), req, &respData); err != nil {
		fmt.Println(err)
		fmt.Println("Error in finding match stats")
		return false
	}

	matchStatsExist := false
	// Check match ID + map number to NOT exist in order to push new stats
	for _, matchStat := range respData.Match.Stats {
		if respData.Match.Id == coreID && matchStat.MapNumber == mapNumber {
			matchStatsExist = true
		}
	}

	if matchStatsExist {
		fmt.Println("MATCH ALREADY HAS STATS!!")
	}

	return !matchStatsExist
}

func addMatchStats(token string, game *game) {
	fmt.Println("ADDING MATCH")

	client := graphql.NewClient(os.Getenv("GRAPHQL_ENDPOINT"))

	req := graphql.NewRequest(`
		mutation AddMatchStats($matchId: String!, $score: ScoreInput!, $matchStatsInput: MatchStatsInput!, $mapName: String!, $mapNumber: Int!) {
			addMatchStats(matchId: $matchId, score: $score, matchStatsInput: $matchStatsInput, mapName: $mapName, mapNumber: $mapNumber) {
				id
			}
		}
	`)

	matchStatsInput := Dictionary{
		"playerStats": []Dictionary{},
		"rounds":      []Dictionary{},
		"teamStats":   []Dictionary{},
	}

	for _, player := range game.totalPlayerStats {
		playerStat := getPlayerAPIDict("b", player)

		matchStatsInput["playerStats"] = append(matchStatsInput["playerStats"].([]Dictionary), playerStat)
	}

	for _, player := range game.tPlayerStats {
		playerStat := getPlayerAPIDict("t", player)

		matchStatsInput["playerStats"] = append(matchStatsInput["playerStats"].([]Dictionary), playerStat)
	}

	for _, player := range game.ctPlayerStats {
		playerStat := getPlayerAPIDict("ct", player)

		matchStatsInput["playerStats"] = append(matchStatsInput["playerStats"].([]Dictionary), playerStat)
	}

	for _, round := range game.rounds {
		roundStat := Dictionary{
			"ctPlayers":           round.initCTerroristCount,
			"tPlayers":            round.initTerroristCount,
			"defuserSteamId":      strconv.FormatUint(round.defuser, 10),
			"planterSteamId":      strconv.FormatUint(round.planter, 10),
			"roundLength":         (round.endingTick - round.startingTick) / game.tickRate,
			"roundNumber":         round.roundNum,
			"roundWinnerTeamName": round.winnerClanName,
			"sideWinner":          strconv.Itoa(round.winnerENUM),
		}

		matchStatsInput["rounds"] = append(matchStatsInput["rounds"].([]Dictionary), roundStat)
	}

	for name, team := range game.totalTeamStats {
		teamStat := Dictionary{
			"teamName":      name,
			"clutches":      team.clutches,
			"deaths":        team.deaths,
			"ef":            team.ef,
			"fa":            team.fass,
			"fourVFives":    team._4v5s,
			"pistolsWon":    team.pistolsW,
			"rounds":        0,
			"roundsAgainst": 0,
			"roundsWon":     0,
			"saves":         0,
			"side":          "t",
			"traded":        team.traded,
			"ud":            team.ud,
			"util":          team.util,
			"wonFourVFives": team._4v5w,
		}

		matchStatsInput["teamStats"] = append(matchStatsInput["teamStats"].([]Dictionary), teamStat)
	}

	score := Dictionary{
		"score":      game.result,
		"team1Name":  game.teams[game.teamOrder[0]].name,
		"team2Name":  game.teams[game.teamOrder[1]].name,
		"team1Score": game.teams[game.teamOrder[0]].score,
		"team2Score": game.teams[game.teamOrder[1]].score,
	}

	req.Var("matchId", game.coreID)
	req.Var("score", score)
	req.Var("mapName", game.mapName)
	req.Var("mapNumber", game.mapNum)
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
