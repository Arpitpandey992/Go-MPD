package server

import (
	"fmt"
	"strings"

	"github.com/arpitpandey992/go-mpd/internal/database"
)

type DbRequestsHandler struct {
	database *database.AudioMeilisearchClient
}

func getNewDbRequestsHandler(db *database.AudioMeilisearchClient) *DbRequestsHandler {
	return &DbRequestsHandler{
		database: db,
	}
}

func (drh *DbRequestsHandler) HandleDbRequest(commands []string) (string, error) {
	mainCommand := strings.ToLower(commands[0])
	switch mainCommand {
	case "search":
		if len(commands) < 2 {
			return "", fmt.Errorf("add: search term missing, expected 1 arg, got 0") // TODO: move this argument parsing logic to a separate centralized module
		}
		return drh.searchInDb(commands[1])
	// case "play":
	// 	return arh.playCurrentTrackInQueue()
	// case "pause":
	// 	return arh.pauseCurrentlyPlayingTrack()
	// case "seek":
	// 	if len(commands) < 2 {
	// 		return "", fmt.Errorf("seek: duration missing, expected 1 arg, got 0")
	// 	}
	// 	return arh.seekCurrentlyPlayingTrack(commands[1])
	// case "next":
	// 	return arh.next()
	// case "prev":
	// 	return arh.previous()
	// case "stop":
	// 	return arh.stopQueuePlayback()
	default:
		return "", fmt.Errorf("unknown database command: %s", mainCommand)
	}
}

func (drh *DbRequestsHandler) searchInDb(searchTerm string) (string, error) {
	results, err := drh.database.SearchAudioFiles(searchTerm, 20) // TODO: do something about this limit variable. Possibly allow arguments like --limit=20 in database commands
	if err != nil {
		return "", err
	}
	filePaths := []string{}
	for _, result := range results {
		filePaths = append(filePaths, result.FilePath)
	}
	return strings.Join(filePaths, "\n"), nil
}
