package server

import (
	"fmt"
	"log"
	"path"
	"time"

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

func (arh *AudioRequestsHandler) HandleAudioRequest(commands []string) (string, error) {
	mainCommand := commands[0]
	switch mainCommand {
	case "add":
		if len(commands) < 2 {
			return "", fmt.Errorf("add: filepath missing, expected 1 arg, got 0") // TODO: move this argument parsing logic to a separate centralized module
		}
		return arh.addToPlaybackQueue(commands[1])
	case "play":
		return arh.playCurrentTrackInQueue()
	case "pause":
		return arh.pauseCurrentlyPlayingTrack()
	case "seek":
		if len(commands) < 2 {
			return "", fmt.Errorf("seek: duration missing, expected 1 arg, got 0")
		}
		return arh.seekCurrentlyPlayingTrack(commands[1])
	case "stop":
		return arh.stopQueuePlayback()
	default:
		return "", fmt.Errorf("unknown audio playback command: %s", mainCommand)
	}
}

func (arh *AudioRequestsHandler) addToPlaybackQueue(filePath string) (string, error) {
	err := arh.PlaybackManager.AddAudioFileToQueue(filePath)
	if err != nil {
		return "", err
	}
	log.Printf("added audio file at: %s to playback queue", filePath)
	return fmt.Sprintf("added %v to playback queue", path.Base(filePath)), nil
}

func (arh *AudioRequestsHandler) playCurrentTrackInQueue() (string, error) {
	err := arh.PlaybackManager.Play()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Playing: %s", arh.PlaybackManager.GetCurrentTrackName()), nil
}
func (arh *AudioRequestsHandler) pauseCurrentlyPlayingTrack() (string, error) {
	err := arh.PlaybackManager.Pause()
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("Paused: %s", arh.PlaybackManager.GetCurrentTrackName()), nil
}

func (arh *AudioRequestsHandler) seekCurrentlyPlayingTrack(durationString string) (string, error) {
	duration, err := time.ParseDuration(durationString)
	if err != nil {
		return "", err
	}
	err = arh.PlaybackManager.Seek(duration)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Seeked: %s to %v", arh.PlaybackManager.GetCurrentTrackName(), duration), nil
}

func (arh *AudioRequestsHandler) stopQueuePlayback() (string, error) {
	err := arh.PlaybackManager.Stop()
	if err != nil {
		return "", err
	}
	return "Stopped Queue Playback", nil
}
