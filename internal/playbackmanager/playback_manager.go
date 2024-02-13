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
* improve logging to be multilevel
* move pm.playbackQueue[pm.QueuePosition] to a function for getting currently playing track's name (and path as well)
* make add AudioFilesToQueue non blocking
*/

// const (
// 	bufferSize int = 100 // TODO : remove this restriction and try to do this without a buffered channel
// )

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
	// _ = speaker.Init(beep.SampleRate(44100), 0) // Initializing the speaker, resampling must be done after creation of AudioPlayer, TODO: use this later when resampling is implemented
	go playbackManager.waitAndManagePlayback()
	return &playbackManager
}

func (pm *PlaybackManager) AddAudioFilesToQueue(filePaths ...string) error {
	unsuccessfulAdditions := make([]string, 0)
	for _, filePath := range filePaths {
		err := pm.AddAudioFileToQueue(filePath)
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

func (pm *PlaybackManager) AddAudioFileToQueue(filePath string) error {
	err := audioplayer.IsFileSupported(filePath)
	if err != nil {
		return err
	}
	go func() {
		pm.nextTrackAdded <- true
	}() // Trying to simulate an infinite buffer, instead of blocking here, this will send the signal at it's own leisure
	pm.playbackQueue = append(pm.playbackQueue, filePath)
	return nil
}

func (pm *PlaybackManager) Next() {
	if pm.audioPlayer != nil {
		pm.audioPlayer.Close()
		pm.audioPlayer = nil
	}
	pm.QueuePosition++
	if pm.QueuePosition == len(pm.playbackQueue) {
		go func() {
			pm.QueuePlaybackFinished <- true
		}()
		pm.isQueuePaused = true
		log.Print("reached the end of playback queue")
		return
	} else if !pm.isQueuePaused {
		pm.isQueuePaused = true // done to not throw exception on hitting Play(), also, it is accurate that after audioPlayer is closed, the queue is paused for a split second
		go func() {
			_ = pm.Play() // TODO for far future: instead of just invoking Play(), call a separate function which handles the transition between tracks
		}()
	}
}

func (pm *PlaybackManager) Play() error {
	if pm.QueuePosition == len(pm.playbackQueue) {
		return fmt.Errorf("no active audio file in queue")
	}
	if pm.audioPlayer == nil { // TODO: This causes race condition, Protect this section with a Mutex, use double-checked locking for optimization
		<-pm.newAudioPlayerCreated
	}
	if !pm.isQueuePaused {
		return fmt.Errorf("queue is already playing")
	}
	pm.isQueuePaused = false // TODO: Protect this variable with a mutex
	log.Printf("playing: %s", pm.GetCurrentTrackName())
	pm.audioPlayer.Play()
	return nil
}

func (pm *PlaybackManager) Pause() error {
	if pm.QueuePosition == len(pm.playbackQueue) || pm.audioPlayer == nil {
		return fmt.Errorf("no active audio file in queue")
	}
	if pm.isQueuePaused {
		return fmt.Errorf("queue is already paused")
	}
	pm.isQueuePaused = true
	pm.audioPlayer.Pause()
	return nil
}

func (pm *PlaybackManager) Stop() error {
	if pm.QueuePosition == len(pm.playbackQueue) || pm.audioPlayer == nil {
		return fmt.Errorf("no active audio file in queue")
	}
	pm.audioPlayer.Stop()
	return nil
}

func (pm *PlaybackManager) Seek(seekTime time.Duration) error {
	if pm.QueuePosition == len(pm.playbackQueue) || pm.audioPlayer == nil {
		return fmt.Errorf("no active audio file in queue")
	}
	return pm.audioPlayer.Seek(seekTime)
}

// Private Functions
func (pm *PlaybackManager) waitAndManagePlayback() {
	for {
		<-pm.nextTrackAdded // creation of new AudioPlayer is blocked till a new track is added

		log.Printf("creating audioplayer for %s", pm.GetCurrentTrackName())
		err := pm.createAudioPlayerForCurrentTrack()
		if err != nil {
			log.Printf("error while creating audio player for %s [skipped], error: %v", pm.GetCurrentTrackName(), err)
			pm.Next()
		} else {
			log.Printf("waiting for %s to finish playing", pm.GetCurrentTrackName())
			<-pm.trackPlaybackFinished // blocking till playback is finished
		}
	}
}

func (pm *PlaybackManager) createAudioPlayerForCurrentTrack() error {
	doOnFinishPlaying := func() {
		log.Printf("%s finished playing", pm.GetCurrentTrackName())
		pm.Next()
		go func() {
			pm.trackPlaybackFinished <- true
		}()
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
	go func() {
		pm.newAudioPlayerCreated <- true
	}()
	log.Print("new audioplayer created successfully")
	return nil
}

func (pm *PlaybackManager) initSpeakerOrResample() error {
	// TODO: Init at the start only, then Resample every time
	// make sure this init call only happens if the new samplerate is different from the earlier playing samplerate
	// resample should be used here, but there is a quality concern as it resamples it
	err := speaker.Init(pm.audioPlayer.Format.SampleRate, pm.audioPlayer.Format.SampleRate.N(time.Second/10))
	if err != nil {
		return err
	}
	speaker.Play(pm.audioPlayer.Ctrl)
	return nil
}

func (pm *PlaybackManager) GetCurrentTrackName() string {
	if pm.QueuePosition >= 0 && pm.QueuePosition < len(pm.playbackQueue) {
		filePath := pm.playbackQueue[pm.QueuePosition]
		return path.Base(filePath)
	}
	return "[Nothing in Queue]"
}
