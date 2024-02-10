package playbackmanager

import (
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"github.com/arpitpandey992/go-mpd/internal/audioplayer"
	"github.com/gopxl/beep/speaker"
)

/*
TODOs
look at individual TODO marked around code
* improve logging
* move pm.playbackQueue[pm.QueuePosition] to a function for getting currently playing track's name (and path as well)
*/

const (
	bufferSize int = 100 // TODO : remove this restriction and try to do this without a buffered channel
)

type PlaybackManager struct {
	// Externally visible Variables
	QueuePosition int

	// External Channels for synchronization
	QueuePlaybackFinished chan bool

	// Internal Variables
	audioPlayer   *audioplayer.AudioPlayer
	playbackQueue []string // TODO -> move from slice of string to a slice of some interface object for flexibility
	isQueuePaused bool

	// Internal Channels for synchronization
	nextTrackAdded        chan bool
	trackPlaybackFinished chan bool
	newAudioPlayerCreated chan bool
}

func CreatePlaybackManager() *PlaybackManager {
	playbackManager := PlaybackManager{
		QueuePosition:         0,
		QueuePlaybackFinished: make(chan bool),
		playbackQueue:         []string{},
		audioPlayer:           nil,
		isQueuePaused:         true,
		nextTrackAdded:        make(chan bool),
		trackPlaybackFinished: make(chan bool),
		newAudioPlayerCreated: make(chan bool),
	}
	go playbackManager.waitAndManagePlayback()
	return &playbackManager
}

func (pm *PlaybackManager) AddAudioFilesToQueue(filePaths ...string) error {
	unsuccessfulAdditions := make([]string, 0)
	for _, filePath := range filePaths {
		err := pm.addAudioFileToQueue(filePath)
		if err != nil {
			log.Printf("could not add: %s, error: %v", filePath, err)
			unsuccessfulAdditions = append(unsuccessfulAdditions, filePath)
		} else {
			log.Printf("added: %s", filePath)
		}
	}
	allFilesAddedSuccessfully := len(unsuccessfulAdditions) == 0
	if !allFilesAddedSuccessfully {
		return fmt.Errorf("could not add: %s", strings.Join(unsuccessfulAdditions, ","))
	}
	return nil
}

func (pm *PlaybackManager) Next() {
	if pm.audioPlayer != nil {
		pm.audioPlayer.Close()
		pm.audioPlayer = nil
	}
	if pm.QueuePosition >= len(pm.playbackQueue) {
		pm.QueuePlaybackFinished <- true
		log.Print("reached the end of playback queue")
	}
	pm.QueuePosition++
}

func (pm *PlaybackManager) Play() error {
	if pm.audioPlayer == nil {
		<-pm.newAudioPlayerCreated
	}
	if !pm.audioPlayer.IsPaused() {
		return fmt.Errorf("%s is already playing", pm.currentTrackName())
	}
	pm.isQueuePaused = false
	log.Printf("playing: %s", pm.currentTrackName())
	pm.audioPlayer.Play()
	return nil
}

func (pm *PlaybackManager) Pause() error {
	if pm.QueuePosition >= len(pm.playbackQueue) || pm.audioPlayer == nil || pm.audioPlayer.IsPaused() {
		return fmt.Errorf("no song is playing")
	}
	pm.isQueuePaused = true
	pm.audioPlayer.Pause()
	return nil
}

func (pm *PlaybackManager) waitAndManagePlayback() {
	for {
		<-pm.nextTrackAdded // creation of new AudioPlayer is blocked till a new track is added

		log.Printf("creating audioplayer for %s", pm.currentTrackName())
		err := pm.createAudioPlayerForCurrentTrack()
		if err != nil {
			log.Printf("error while creating audio player for %s [skipped], error: %v", pm.currentTrackName(), err)
			pm.Next()
		} else {
			log.Printf("waiting for %s to finish playing", pm.currentTrackName())
			<-pm.trackPlaybackFinished // blocking till playback is finished
		}
	}
}

func (pm *PlaybackManager) addAudioFileToQueue(filePath string) error {
	err := audioplayer.IsFileSupported(filePath)
	if err != nil {
		return err
	}
	go func(pm *PlaybackManager) {
		pm.nextTrackAdded <- true
	}(pm) // Trying to simulate an infinite buffer
	pm.playbackQueue = append(pm.playbackQueue, filePath)
	return nil
}

func (pm *PlaybackManager) createAudioPlayerForCurrentTrack() error {
	doOnFinishPlaying := func() {
		log.Printf("%s finished playing", pm.currentTrackName())
		pm.Next()
		pm.trackPlaybackFinished <- true
	}
	startPaused := pm.isQueuePaused
	ap, err := audioplayer.CreateAudioPlayer(pm.playbackQueue[pm.QueuePosition], doOnFinishPlaying, startPaused)
	if err != nil {
		return err
	}
	pm.audioPlayer = ap
	err = pm.initSpeakerOrResample()
	if err != nil {
		return err
	}
	go func(pm *PlaybackManager) {
		pm.newAudioPlayerCreated <- true
	}(pm)
	log.Print("new audioplayer created successfully")
	return nil
}

func (pm *PlaybackManager) initSpeakerOrResample() error {
	// make sure this init call only happens if the new samplerate is different from the earlier playing samplerate
	// resample should be used here, but there is a quality concern as it resamples it
	err := speaker.Init(pm.audioPlayer.Format.SampleRate, pm.audioPlayer.Format.SampleRate.N(time.Second/10)) // TODO: use Resample instead
	if err != nil {
		return err
	}
	speaker.Play(pm.audioPlayer.Ctrl)
	return nil
}

func (pm *PlaybackManager) currentTrackName() string {
	if pm.QueuePosition >= 0 && pm.QueuePosition < len(pm.playbackQueue) {
		filePath := pm.playbackQueue[pm.QueuePosition]
		return path.Base(filePath)
	}
	return "[Nothing in Queue]"
}
