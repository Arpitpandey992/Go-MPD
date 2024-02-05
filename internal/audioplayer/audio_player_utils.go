package audioplayer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
)

func logError(err error) error {
	if err != nil {
		log.Printf("Error: %s", err)
	}
	return err
}

func IsFileSupported(filePath string) error {
	supportedExtensions := []string{".mp3", ".flac"}
	fileExtension := strings.ToLower(filepath.Ext(filePath))
	for _, extension := range supportedExtensions {
		if extension == fileExtension {
			return nil
		}
	}
	return fmt.Errorf("audio format: %s is not supported yet", fileExtension)
}

func CreateAudioPlayer(filePath string, callbackFunction func()) (*AudioPlayer, error) {
	err := IsFileSupported(filePath)
	if logError(err) != nil {
		return nil, err
	}
	audioFile, err := os.Open(filePath)
	if logError(err) != nil {
		return nil, err
	}
	defer audioFile.Close()
	audioPlayer, err := getNewAudioPlayer(audioFile, callbackFunction)
	if logError(err) != nil {
		return nil, err
	}
	return audioPlayer, nil
}

func getNewAudioPlayer(file *os.File, callbackfunc func()) (*AudioPlayer, error) {
	fileFormat := strings.ToLower(filepath.Ext(file.Name()))
	var DecoderMap = map[string]func(*os.File) (s beep.StreamSeekCloser, format beep.Format, err error){
		".mp3":  func(file *os.File) (s beep.StreamSeekCloser, format beep.Format, err error) { return mp3.Decode(file) },
		".flac": func(file *os.File) (s beep.StreamSeekCloser, format beep.Format, err error) { return flac.Decode(file) },
	}
	streamer, format, err := DecoderMap[fileFormat](file)
	if logError(err) != nil {
		return nil, err
	}
	ctrl := &beep.Ctrl{Streamer: beep.Seq(streamer, beep.Callback(callbackfunc)), Paused: true}
	return &AudioPlayer{Ctrl: ctrl, streamer: streamer, Format: format}, nil
}
