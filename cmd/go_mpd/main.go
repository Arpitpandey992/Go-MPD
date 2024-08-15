package main

import (
	"log"

	"github.com/arpitpandey992/go-mpd/internal/config"
	"github.com/arpitpandey992/go-mpd/internal/database"
	"github.com/arpitpandey992/go-mpd/internal/server"
)

func main() {
	config, err := config.GetBaseConfiguration()
	if err != nil {
		log.Fatal("could get configuration", err)
	}
	audioMeilisearchClient := database.GetNewAudioMeiliSearchClient(config)
	// database.SearchWithUserInput(audioMeilisearchClient)
	server.CreateAndStartServer(audioMeilisearchClient)
	select {}
}
