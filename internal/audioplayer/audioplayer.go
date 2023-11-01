package audioplayer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func logError(err error) error {
	if err != nil {
		log.Printf("Error: %s", err)
	}
	return err
}

func CreateAudioPlayer(filePath string) (*AudioPlayer, error) {
	audioFile, err := os.Open(filePath)
	if logError(err) != nil {
		return nil, err
	}
	audioPlayer, err := getNewAudioPlayer(audioFile)
	if logError(err) != nil {
		return nil, err
	}
	return audioPlayer, nil
}

func getNewAudioPlayer(file *os.File) (*AudioPlayer, error) {
	fileFormat := strings.ToLower(filepath.Ext(file.Name()))
	if _, exists := DecoderMap[fileFormat]; exists {
		streamer, format, err := DecoderMap[fileFormat](file)
		if logError(err) != nil {
			return nil, err
		}
		return &AudioPlayer{streamer: streamer, format: format}, nil
	}
	return nil, fmt.Errorf("audio format: %s is not supported yet", fileFormat)
}
