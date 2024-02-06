package server

import (
	"fmt"
	"log"

	"github.com/arpitpandey992/go-mpd/internal/playbackmanager"
)

type AudioRequestsHandler struct {
	PlaybackManager *playbackmanager.PlaybackManager
}

func getNewAudioRequestsHandler() *AudioRequestsHandler {
	return &AudioRequestsHandler{
		PlaybackManager: playbackmanager.CreatePlaybackManager(),
	}
}

func (arh *AudioRequestsHandler) HandleAudioRequest(commands []string) error {
	mainCommand := commands[0]
	switch mainCommand {
	case "add":
		return arh.AddToPlaybackQueue(commands[1])
	case "play":
		return arh.PlayCurrentTrackInQueue()
	case "pause":
		return arh.PauseCurrentlyPlayingTrack()
	default:
		return fmt.Errorf("unknown command: %s", mainCommand)
	}
}

func (arh *AudioRequestsHandler) AddToPlaybackQueue(filePath string) error {
	err := arh.PlaybackManager.AddAudioFilesToQueue(filePath)
	if err != nil {
		return err
	}
	log.Printf("added audio file at: %s to playback queue", filePath)
	return nil
}

func (arh *AudioRequestsHandler) PlayCurrentTrackInQueue() error {
	return arh.PlaybackManager.Play()
}
func (arh *AudioRequestsHandler) PauseCurrentlyPlayingTrack() error {
	return arh.PlaybackManager.Pause()
}
