package audioplayer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
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

var DecoderMap = map[string]func(*os.File) (s beep.StreamSeekCloser, format beep.Format, err error){
	".mp3":  func(file *os.File) (s beep.StreamSeekCloser, format beep.Format, err error) { return mp3.Decode(file) },
	".flac": func(file *os.File) (s beep.StreamSeekCloser, format beep.Format, err error) { return flac.Decode(file) },
}

type AudioPlayer struct {
	streamer beep.StreamSeekCloser
	format   beep.Format
}

func (ap *AudioPlayer) Play() error {
	// make sure this init call only happens if the new samplerate is different from the earlier playing samplerate
	// resample should be used here, but there is a quality concern as it resamples it
	err := speaker.Init(ap.format.SampleRate, ap.format.SampleRate.N(time.Second/10))
	if err != nil {
		return err
	}
	speaker.Play(ap.streamer)
	return nil
}

func (ap *AudioPlayer) Close() error {
	return ap.streamer.Close()
}
