package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hasura/go-graphql-client"
	"github.com/joho/godotenv"
)

func authenticate() graphql.Client {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	graphqlEndpoint := os.Getenv("GRAPHQL_ENDPOINT")
	graphqlDiscord := os.Getenv("GRAPHQL_DISCORDID")
	graphqlPass := os.Getenv("GRAPHQL_PASSWORD")

	client := graphql.NewClient(graphqlEndpoint, nil)

	var m struct {
		TokenAuth struct {
			Token graphql.String
		} `graphql:"tokenAuth(discordId: $discordId, password: $apiKey)"`
	}

	variables := map[string]interface{}{
		"discordId": graphql.String(graphqlDiscord),
		"apiKey":    graphql.String(graphqlPass),
	}

	err2 := client.Mutate(context.Background(), &m, variables)
	if err2 != nil {
		// Handle error.
		fmt.Println(err2)
	}

	addAuthHeader := func(request *http.Request) {
		request.Header.Add("Authorization", "JWT "+string(m.TokenAuth.Token))
	}

	return *graphql.NewClient(graphqlEndpoint, nil).WithRequestModifier(addAuthHeader)

	//return m.TokenAuth.Token
}

func verifyOriginalMatch(client graphql.Client, coreID string) bool {
	var q struct {
		Match struct {
			Id    graphql.ID
			Stats struct {
				Id graphql.ID
			}
		} `graphql:"match(matchId: $coreID)"`
	}

	variables := map[string]interface{}{
		"coreID": graphql.String(coreID),
	}

	err2 := client.Query(context.Background(), &q, variables)
	if err2 != nil {
		// Handle error.
		fmt.Println(err2)
	}

	fmt.Println(q.Match)

	// I think this will explode!
	return q.Match.Id != nil && q.Match.Stats.Id == nil
}

func addMatch(client graphql.Client, game *game) {
	fmt.Println("ADDING MATCH")

	var m struct {
		MatchStats struct {
			Id graphql.ID
		} `graphql:"addMatchStats(matchId: $matchId, matchStatsInput: $matchStatsInput)"`
	}

	type TeamStatsInput struct {
		teamName graphql.String
	}

	type PlayerStatsInput struct {
		playerSteamId graphql.String
	}

	type RoundsStatsInput struct {
		roundNumber graphql.Int
	}

	type MatchStatsInput struct {
		teamStats   [1]*TeamStatsInput
		playerStats []*PlayerStatsInput
		rounds      []*RoundsStatsInput
	}

	teamStatsInput := TeamStatsInput{teamName: graphql.String("Team A")}

	matchStatsInput := MatchStatsInput{}

	matchStatsInput.teamStats[0] = &teamStatsInput

	variables := map[string]interface{}{
		"matchId":         graphql.String(game.coreID),
		"matchStatsInput": matchStatsInput,
	}

	fmt.Println(variables)

	err2 := client.Mutate(context.Background(), &m, variables)
	if err2 != nil {
		// Handle error.
		fmt.Println(err2)
	}

}
