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
		mutation AddMatchStats($matchId: String!, $matchStatsInput: MatchStatsInput!) {
			addMatchStats(matchId: $matchId, matchStatsInput: $matchStatsInput) {
				id
			}
		}
	`)

	type Dictionary map[string]interface{}

	matchStatsInput := Dictionary{
		"playerStats": []Dictionary{
		},
		"rounds": []Dictionary{
			{},
			{},
		},
		"teamStats": []Dictionary{
			{},
			{},
		},
	}

	for _, player := range game.totalPlayerStats {
		playerStat := Dictionary{
			"playerSteamId": strconv.FormatUint(player.steamID, 10),
			"side": strconv.Itoa(player.side),
			"adp": player.tADP,
			"adr": player.tADR,
		}

		matchStatsInput["playerStats"] = append(matchStatsInput["playerStats"].([]Dictionary), playerStat)
	}

	req.Var("matchId", game.coreID)
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
