package server

import (
	"fmt"
	"github.com/arpitpandey992/go-mpd/internal/audioplayer"
	"log"
	"os"
)

func HandleAudioRequest(commands []string) error {
	mainCommand := commands[0]
	if mainCommand == "play" {
		err := playAudioFile(commands[1])
		if err != nil {
			return err
		}
	}
	return fmt.Errorf("unknown command: %s", mainCommand)
}

func playAudioFile(filePath string) error {
	if !doesFileExists(filePath) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}
	log.Printf("playing audio file at: %s", filePath)
	audioPlayer, err := audioplayer.CreateAudioPlayer(filePath)
	if err != nil {
		return err
	}
	err = audioPlayer.Play()
	return err
}

func doesFileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
