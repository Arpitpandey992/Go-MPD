package playbackmanager

import (
	"fmt"
	"log"
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
	QueueBufferSize = 100
)

type PlaybackManager struct {
	// External Variables
	QueuePosition int

	// External Channels for synchronization
	QueuePlaybackFinished chan bool

	// Internal Variables
	audioPlayer   *audioplayer.AudioPlayer
	playbackQueue []string // TODO -> move from slice of string to a slice of some interface object for flexibility

	// Internal Channels for synchronization
	newTrackAdded         chan bool
	trackPlaybackFinished chan bool
	newAudioPlayerCreated chan bool
}

func CreatePlaybackManager() *PlaybackManager {
	playbackManager := PlaybackManager{
		QueuePosition:         0,
		QueuePlaybackFinished: make(chan bool),
		playbackQueue:         []string{},
		audioPlayer:           nil,
		newTrackAdded:         make(chan bool, QueueBufferSize),
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

func (pm *PlaybackManager) addAudioFileToQueue(filePath string) error {
	err := audioplayer.IsFileSupported(filePath)
	if err != nil {
		return err
	}
	pm.playbackQueue = append(pm.playbackQueue, filePath)
	pm.newTrackAdded <- true //signal the channel to proceed with the creation of AudioPlayer for the new track if it is waiting
	return nil
}

func (pm *PlaybackManager) Play() error {
	if pm.QueuePosition >= len(pm.playbackQueue) {
		pm.QueuePlaybackFinished <- true
		return fmt.Errorf("reached the end of playback queue")
	}
	if pm.audioPlayer == nil {
		// waiting for the audioplayer of the current track to be created
		<-pm.newAudioPlayerCreated
	}
	if !pm.audioPlayer.IsPaused() {
		return fmt.Errorf("%s is already playing", pm.playbackQueue[pm.QueuePosition])
	}
	log.Printf("playing: %s", pm.playbackQueue[pm.QueuePosition])
	pm.audioPlayer.Play()
	return nil
}

func (pm *PlaybackManager) Pause() error {
	if pm.QueuePosition >= len(pm.playbackQueue) || pm.audioPlayer == nil || pm.audioPlayer.IsPaused() {
		return fmt.Errorf("no song is playing")
	}
	pm.audioPlayer.Pause()
	return nil
}

func (pm *PlaybackManager) waitAndManagePlayback() {
	for {
		<-pm.newTrackAdded // creation of new AudioPlayer is blocked till a new track is added
		log.Printf("creating audioplayer for %s", pm.playbackQueue[pm.QueuePosition])
		err := pm.createAudioPlayerForCurrentTrack()
		if err != nil {
			log.Printf("error while creating audio player for %s [skipped], error: %v", pm.playbackQueue[pm.QueuePosition], err)
		} else {
			<-pm.trackPlaybackFinished // blocking till playback is finished
			log.Printf("%s finished playing", pm.playbackQueue[pm.QueuePosition])
			go pm.doOnTrackPlaybackFinished()
		}
	}
}

func (pm *PlaybackManager) createAudioPlayerForCurrentTrack() error {
	doOnFinishPlaying := func() {
		pm.trackPlaybackFinished <- true
	}
	ap, err := audioplayer.CreateAudioPlayer(pm.playbackQueue[pm.QueuePosition], doOnFinishPlaying)
	if err != nil {
		return err
	}
	pm.audioPlayer = ap
	err = pm.initSpeakerOrResample()
	if err != nil {
		return err
	}
	pm.newAudioPlayerCreated <- true
	return nil
}

func (pm *PlaybackManager) initSpeakerOrResample() error {
	// make sure this init call only happens if the new samplerate is different from the earlier playing samplerate
	// resample should be used here, but there is a quality concern as it resamples it
	err := speaker.Init(pm.audioPlayer.Format.SampleRate, pm.audioPlayer.Format.SampleRate.N(time.Second/10))
	if err != nil {
		return err
	}
	speaker.Play(pm.audioPlayer.Ctrl)
	return nil
}

func (pm *PlaybackManager) doOnTrackPlaybackFinished() {
	pm.audioPlayer.Close()
	pm.audioPlayer = nil
	pm.QueuePosition++
	pm.Play() // if there is a new track available, play it, otherwise signals the QueuePlaybackFinished
}
