package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	//dem "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
)

func beginOutput(game *game) {

	setDisplayOrders(game)

	m_ID := createHash(game)
	fmt.Println("M_ID", m_ID)

	outputFile, outputFileErr := os.Create("out/gameStats.csv")
	if outputFileErr != nil {
		fmt.Println("OH NOE")
	}
	w := csv.NewWriter(outputFile)

	records := [][]string{
		{"m_ID", "Map", "Team", "Name", "Rating"},
	}

	teamA := []string{m_ID, game.mapName, "", game.teams[game.teamOrder[0]].name}
	records = append(records, [][]string{teamA}...)

	for _, steam := range game.playerOrder {
		player := game.totalPlayerStats[steam]
		if player.teamID == game.teamOrder[0] {
			playerOutput := []string{m_ID, game.mapName, game.teams[game.teamOrder[0]].name, player.name, fmt.Sprintf("%.2f", player.rating)}
			records = append(records, [][]string{playerOutput}...)
		}
	}

	teamB := []string{m_ID, game.mapName, "", game.teams[game.teamOrder[1]].name}
	records = append(records, [][]string{teamB}...)

	for _, steam := range game.playerOrder {
		player := game.totalPlayerStats[steam]
		if player.teamID == game.teamOrder[1] {
			playerOutput := []string{m_ID, game.mapName, game.teams[game.teamOrder[1]].name, player.name, fmt.Sprintf("%.2f", player.rating)}
			records = append(records, [][]string{playerOutput}...)
		}
	}

	result := "" + game.teams[game.teamOrder[0]].name + " " + strconv.Itoa(game.teams[game.teamOrder[0]].score) + " - " + strconv.Itoa(game.teams[game.teamOrder[1]].score) + " " + game.teams[game.teamOrder[1]].name
	resultLine := []string{"1", game.mapName, result}
	records = append(records, [][]string{resultLine}...)

	for i, _ := range records {
		w.Write(records[i])
	}
	w.Flush()
}

func createHash(game *game) string {
	fmt.Println("headerTickL", game.tickLength)
	hashValue := fmt.Sprint(game.tickLength)
	totalDamage := 0
	totalUD := 0
	playerInitial := ""

	for _, player := range game.totalPlayerStats {
		totalDamage += player.damage
		totalUD += player.utilDmg
		playerInitial += string(player.name[0])
	}

	s := strings.Split(playerInitial, "")
	sort.Strings(s)
	playerInitial = strings.Join(s, "")

	fmt.Println("tick", hashValue)
	hashValue += fmt.Sprint(totalDamage) + playerInitial

	return randomizeHash(hashValue, totalUD)
}

func randomizeHash(hashValue string, seedVal int) string {
	rand.Seed(int64(seedVal))

	hashValueRune := []rune(hashValue)
	rand.Shuffle(len(hashValueRune), func(i, j int) {
		hashValueRune[i], hashValueRune[j] = hashValueRune[j], hashValueRune[i]
	})

	return string(hashValueRune)
}

func setDisplayOrders(game *game) {
	if game.winnerID != 0 {
		game.teamOrder = append(game.teamOrder, game.winnerID)
		for teamID, _ := range game.teams {
			if !intInSlice(teamID, game.teamOrder) {
				game.teamOrder = append(game.teamOrder, teamID)
			}
		}
	} else {
		//just sort alphabetically
		for teamID, _ := range game.teams {
			if len(game.teamOrder) == 0 {
				game.teamOrder = append(game.teamOrder, teamID)
			} else {
				if game.teams[game.teamOrder[0]].name < game.teams[teamID].name {
					game.teamOrder = append(game.teamOrder, teamID)
				} else {
					game.teamOrder = append(game.teamOrder, game.teamOrder[0])
					game.teamOrder[0] = teamID
				}
			}
		}
	}

	for _, teamID := range game.teamOrder {
		offset := len(game.playerOrder)
		for steam, player := range game.totalPlayerStats {
			if player.teamID == teamID {
				if len(game.playerOrder) > offset {
					//subsetI := len(game.playerOrder) - offset
					for index, _ := range game.playerOrder[offset:] {
						if player.rating > game.totalPlayerStats[game.playerOrder[index+offset]].rating {
							game.playerOrder = append(game.playerOrder[:index+offset+1], game.playerOrder[index+offset:]...)
							game.playerOrder[index+offset] = steam
							break
						} else if (index+offset)+1 == len(game.playerOrder) {
							game.playerOrder = append(game.playerOrder, steam)
							break
						} else {
							continue
						}
					}
				} else {
					game.playerOrder = append(game.playerOrder, steam)
				}
			}
		}
	}
	fmt.Println(game.teamOrder)
	fmt.Println(game.playerOrder)
}
