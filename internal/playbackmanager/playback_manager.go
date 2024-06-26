package playbackmanager

import (
	"fmt"
	"log"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/arpitpandey992/go-mpd/internal/audioplayer"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
)

/*
TODOs
look at individual TODO marked around code
* improve logging to be multilevel
* move pm.playbackQueue[pm.QueuePosition] to a function for getting currently playing track's name (and path as well)
* make add AudioFilesToQueue non blocking
* move the initSpeaker logic to audio_player later to fully decouple beep from playback_manager
* Then convert the AudioPlayer to an interface where we can later use oto as well, BeepAudioPlayer can implement AudioPlayer
* Move the conditional variable Mutex pair to a Struct
*/

const (
	baseSampleRate = 44100 //TODO: make this a quality setting instead since this would mean we are gonna downsample higher sample rate files to 44100 only
)

type PlaybackManager struct {
	QueuePosition int

	QueuePlaybackFinished chan bool

	audioPlayer   *audioplayer.AudioPlayer
	playbackQueue []string // TODO -> move from slice of string to a slice of some interface object for flexibility

	audioPlayerLock   sync.Mutex
	playbackQueueLock sync.Mutex
}

func CreatePlaybackManager() *PlaybackManager {
	playbackManager := PlaybackManager{
		QueuePosition:         0,
		QueuePlaybackFinished: make(chan bool),
		playbackQueue:         []string{},
		audioPlayer:           nil,
		audioPlayerLock:       sync.Mutex{},
		playbackQueueLock:     sync.Mutex{},
	}
	speakerSampleRate := beep.SampleRate(baseSampleRate)
	_ = speaker.Init(speakerSampleRate, speakerSampleRate.N(time.Second/10))
	return &playbackManager
}

func (pm *PlaybackManager) AddAudioFilesToQueue(filePaths ...string) error {
	pm.playbackQueueLock.Lock()
	defer pm.playbackQueueLock.Unlock()
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
	log.Print("added all files")
	return nil
}

func (pm *PlaybackManager) Next() error {
	log.Printf("debug: Next() called")
	pm.audioPlayerLock.Lock()
	pm.playbackQueueLock.Lock()
	defer pm.audioPlayerLock.Unlock()
	defer pm.playbackQueueLock.Unlock()
	return pm.next()
}

func (pm *PlaybackManager) Previous() error {
	log.Printf("debug: Previous() called")
	pm.audioPlayerLock.Lock()
	pm.playbackQueueLock.Lock()
	defer pm.audioPlayerLock.Unlock()
	defer pm.playbackQueueLock.Unlock()
	return pm.previous()
}

func (pm *PlaybackManager) Play() error {
	pm.audioPlayerLock.Lock()
	defer pm.audioPlayerLock.Unlock()
	return pm.play()
}

func (pm *PlaybackManager) Pause() error {
	pm.audioPlayerLock.Lock()
	pm.playbackQueueLock.Lock()
	defer pm.audioPlayerLock.Unlock()
	defer pm.playbackQueueLock.Unlock()
	return pm.pause()
}

func (pm *PlaybackManager) Stop() error {
	pm.audioPlayerLock.Lock()
	pm.playbackQueueLock.Lock()
	defer pm.audioPlayerLock.Unlock()
	defer pm.playbackQueueLock.Unlock()
	return pm.stop()
}

func (pm *PlaybackManager) Seek(seekTime time.Duration) error {
	pm.audioPlayerLock.Lock()
	pm.playbackQueueLock.Lock()
	defer pm.audioPlayerLock.Unlock()
	defer pm.playbackQueueLock.Unlock()

	if pm.QueuePosition == len(pm.playbackQueue) || pm.audioPlayer == nil {
		return fmt.Errorf("no active audio file in queue")
	}
	return pm.audioPlayer.Seek(seekTime)
}

func (pm *PlaybackManager) GetCurrentTrackName() string {
	if pm.QueuePosition >= 0 && pm.QueuePosition < len(pm.playbackQueue) {
		filePath := pm.playbackQueue[pm.QueuePosition]
		return path.Base(filePath)
	}
	return "[Nothing in Queue]"
}

func (pm *PlaybackManager) createAudioPlayerForCurrentTrack() error {
	log.Printf("creating audioplayer for: %s", pm.GetCurrentTrackName())
	if pm.audioPlayer != nil {
		return nil
	}
	doOnFinishPlaying := func() {
		// This is ran on a separate go routine because the speaker is locked when the callback function is called. Hence, it causes deadlock as Next() also requires the speaker to be locked.
		go func() {
			err := pm.Next()
			if err != nil {
				log.Print(err)
			}
		}()
	}
	ap, err := audioplayer.CreateAudioPlayer(pm.playbackQueue[pm.QueuePosition], doOnFinishPlaying)
	if err != nil {
		return err
	}
	pm.audioPlayer = ap
	pm.resampleTrackIfNeeded()
	log.Print("new audioplayer created successfully")
	return nil
}

func (pm *PlaybackManager) resampleTrackIfNeeded() {
	newSampleRate := pm.audioPlayer.Format.SampleRate
	if int32(newSampleRate) != baseSampleRate {
		log.Printf("resampling %s from %d to %d", pm.GetCurrentTrackName(), int32(newSampleRate), baseSampleRate)
		resampled := beep.Resample(4, newSampleRate, beep.SampleRate(baseSampleRate), pm.audioPlayer.Ctrl)
		speaker.Play(resampled)
	} else {
		speaker.Play(pm.audioPlayer.Ctrl)
	}
}

func (pm *PlaybackManager) moveQueuePosition(delta int) error {

	/*
		allows the following final positions:
		position:             0, 1, ..., n-2, n-1, n
		music file available: Y, Y, ..., Y  , Y  , N
		So, the queue position can reach just next to the last available track at which point it should be propagated that Queue is finished
	*/
	finalPosition := pm.QueuePosition + delta
	if finalPosition < 0 || finalPosition > len(pm.playbackQueue) {
		return fmt.Errorf("invalid queue pointer move request: %d", finalPosition)
	}

	err := pm.stop()
	if err != nil {
		return fmt.Errorf("error while stopping current playback: %s", err.Error())
	}
	pm.QueuePosition = finalPosition
	return nil
}

func (pm *PlaybackManager) isQueuePaused() bool {
	return pm.audioPlayer != nil && pm.audioPlayer.IsPaused()
}

func (pm *PlaybackManager) addAudioFileToQueue(filePath string) error {
	err := audioplayer.IsFileSupported(filePath)
	if err != nil {
		return err
	}
	pm.playbackQueue = append(pm.playbackQueue, filePath)
	return nil
}

func (pm *PlaybackManager) play() error {
	if pm.QueuePosition < 0 {
		panic(fmt.Sprintf("Queue position: %d is invalid", pm.QueuePosition))
	}
	if pm.QueuePosition == len(pm.playbackQueue) {
		return fmt.Errorf("no active audio file in queue")
	}
	if pm.audioPlayer == nil {
		err := pm.createAudioPlayerForCurrentTrack()
		if err != nil {
			return err
		}
	}
	if !pm.isQueuePaused() {
		return fmt.Errorf("queue is already playing")
	}
	log.Printf("playing: %s", pm.GetCurrentTrackName())
	pm.audioPlayer.Play()
	return nil
}
func (pm *PlaybackManager) pause() error {
	if pm.QueuePosition == len(pm.playbackQueue) || pm.audioPlayer == nil {
		return fmt.Errorf("no active audio file in queue")
	}
	if pm.isQueuePaused() {
		return fmt.Errorf("queue is already paused")
	}
	pm.audioPlayer.Pause()
	return nil
}

func (pm *PlaybackManager) stop() error {
	if pm.QueuePosition == len(pm.playbackQueue) || pm.audioPlayer == nil {
		return fmt.Errorf("no active audio file in queue")
	}
	err := pm.audioPlayer.Close()
	if err != nil {
		return err
	}
	pm.audioPlayer = nil
	return nil
}

func (pm *PlaybackManager) next() error {
	initiallyQueuePaused := pm.isQueuePaused()
	err := pm.moveQueuePosition(1)
	if err != nil {
		return err
	}
	if pm.QueuePosition == len(pm.playbackQueue) {
		go func() {
			pm.QueuePlaybackFinished <- true
		}()
		log.Print("reached the end of playback queue")
		return nil
	}
	if initiallyQueuePaused {
		return nil
	}
	err = pm.play()
	if err != nil {
		return err
	}
	return nil
}

func (pm *PlaybackManager) previous() error {
	initiallyQueuePaused := pm.isQueuePaused()
	if pm.QueuePosition != 0 {
		err := pm.moveQueuePosition(-1)
		if err != nil {
			return err
		}
	} else {
		return pm.restartTrack()
	}
	if initiallyQueuePaused {
		return nil
	}
	err := pm.play()
	if err != nil {
		return err
	}
	return nil
}

func (pm *PlaybackManager) restartTrack() error {
	err := pm.stop()
	if err != nil {
		return err
	}
	return pm.play()
}
