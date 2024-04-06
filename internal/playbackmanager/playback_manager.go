package playbackmanager

import (
	"fmt"
	"log"
	"path"
	"strings"
	"sync"
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
* move the initSpeaker logic to audio_player later to fully decouple beep from playback_manager
* Then convert the AudioPlayer to an interface where we can later use oto as well, BeepAudioPlayer can implement AudioPlayer
* Move the conditional variable Mutex pair to a Struct
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
	isQueuePaused bool     // TODO -> synchronize this variable

	// Internal variables for synchronization
	trackPlaybackFinished chan bool

	fileReadyForPlaybackConditionalVariable *sync.Cond
	audioPlayerCreationConditionalVariable  *sync.Cond
}

func CreatePlaybackManager() *PlaybackManager {
	playbackManager := PlaybackManager{
		QueuePosition:                           0,
		QueuePlaybackFinished:                   make(chan bool),
		playbackQueue:                           []string{},
		audioPlayer:                             nil,
		isQueuePaused:                           true,
		trackPlaybackFinished:                   make(chan bool),
		fileReadyForPlaybackConditionalVariable: sync.NewCond(&sync.Mutex{}),
		audioPlayerCreationConditionalVariable:  sync.NewCond(&sync.Mutex{}),
	}
	// Initializing Conditional Variables

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
	log.Print("added all files")
	return nil
}

func (pm *PlaybackManager) AddAudioFileToQueue(filePath string) error {
	err := audioplayer.IsFileSupported(filePath)
	if err != nil {
		return err
	}

	pm.playbackQueue = append(pm.playbackQueue, filePath)
	if pm.QueuePosition+1 == len(pm.playbackQueue) {
		pm.fileReadyForPlaybackConditionalVariable.Broadcast()
	}
	return nil
}

func (pm *PlaybackManager) Next() error {
	// _ = pm.Stop()
	log.Printf("debug: Next() called")
	err := pm.moveQueuePosition(1)
	if err != nil {
		return err
	}
	if pm.QueuePosition == len(pm.playbackQueue) {
		go func() {
			pm.QueuePlaybackFinished <- true // TODO: change to a synchronization variable
		}()
		pm.isQueuePaused = true
		log.Print("reached the end of playback queue")
	} else if !pm.isQueuePaused {
		pm.isQueuePaused = true // done to not throw exception on hitting Play(), also, it is accurate that after audioPlayer is closed, the queue is paused for a split second
		go func() {
			_ = pm.Play() // TODO for far future: instead of just invoking Play(), call a separate function which handles the transition between tracks
		}()
	}
	return nil
}

func (pm *PlaybackManager) Previous() {
	err := pm.moveQueuePosition(-1)
	if err != nil {
		log.Print(err)
		return
	}
	if pm.QueuePosition == len(pm.playbackQueue) {
		go func() {
			pm.QueuePlaybackFinished <- true // TODO: change to a synchronization variable
		}()
		pm.isQueuePaused = true
		log.Print("reached the end of playback queue")
	} else if !pm.isQueuePaused {
		pm.isQueuePaused = true // done to not throw exception on hitting Play(), also, it is accurate that after audioPlayer is closed, the queue is paused for a split second
		go func() {
			_ = pm.Play() // TODO for far future: instead of just invoking Play(), call a separate function which handles the transition between tracks
		}()
	}
}

func (pm *PlaybackManager) Play() error {
	log.Print("debug: play() called")
	if pm.QueuePosition == len(pm.playbackQueue) {
		return fmt.Errorf("no active audio file in queue")
	}
	if pm.audioPlayer == nil {
		pm.audioPlayerCreationConditionalVariable.L.Lock()
		if pm.audioPlayer == nil {
			pm.audioPlayerCreationConditionalVariable.Wait()
		}
		pm.audioPlayerCreationConditionalVariable.L.Unlock()
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
	err := pm.Pause()
	if err != nil {
		return err
	}
	return pm.Seek(0)
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
		if pm.QueuePosition == len(pm.playbackQueue) {
			pm.fileReadyForPlaybackConditionalVariable.L.Lock()
			if pm.QueuePosition == len(pm.playbackQueue) {
				pm.fileReadyForPlaybackConditionalVariable.Wait()
			}
			pm.fileReadyForPlaybackConditionalVariable.L.Unlock()
		}

		log.Printf("creating audioplayer for %s", pm.GetCurrentTrackName())
		doOnFinishPlaying := func() {
			log.Printf("%s finished playing", pm.GetCurrentTrackName())
			err := pm.Next()
			if err != nil {
				log.Print(err)
			}
			pm.trackPlaybackFinished <- true
		}
		err := pm.createAudioPlayerForCurrentTrack(doOnFinishPlaying)
		if err != nil {
			log.Printf("error while creating audio player for %s [skipped], error: %v", pm.GetCurrentTrackName(), err)
			err := pm.Next()
			if err != nil {
				log.Print(err)
			}
		} else {
			log.Printf("waiting for %s to finish playing", pm.GetCurrentTrackName())
			<-pm.trackPlaybackFinished // blocking till playback is finished
		}
	}
}

func (pm *PlaybackManager) createAudioPlayerForCurrentTrack(doOnFinishPlaying func()) error {
	pm.audioPlayerCreationConditionalVariable.L.Lock()
	ap, err := audioplayer.CreateAudioPlayer(pm.playbackQueue[pm.QueuePosition], doOnFinishPlaying)
	if err != nil {
		return err
	}
	pm.audioPlayer = ap
	err = pm.initSpeakerOrResample()
	if err != nil {
		return err
	}
	pm.audioPlayerCreationConditionalVariable.Broadcast()
	log.Print("new audioplayer created successfully")
	pm.audioPlayerCreationConditionalVariable.L.Unlock()
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

func (pm *PlaybackManager) moveQueuePosition(delta int) error {
	finalPosition := pm.QueuePosition + delta
	/*
		allows the following final positions:
		position:             0, 1, ..., n-2, n-1, n
		music file available: Y, Y, ..., Y  , Y  , N
		So, the queue position can reach just next to the last available track at which point it should be propagated that Queue is finished
	*/
	if finalPosition < 0 || finalPosition > len(pm.playbackQueue) {
		return fmt.Errorf("invalid queue pointer move request: %d", finalPosition)
	}
	pm.fileReadyForPlaybackConditionalVariable.L.Lock()
	pm.audioPlayerCreationConditionalVariable.L.Lock()
	defer pm.fileReadyForPlaybackConditionalVariable.L.Unlock()
	defer pm.audioPlayerCreationConditionalVariable.L.Unlock()
	// closing the current AudioPlayer
	if pm.audioPlayer != nil {
		pm.audioPlayer.Close()
		pm.audioPlayer = nil
	}
	pm.QueuePosition = finalPosition
	// signal audio file being ready
	pm.fileReadyForPlaybackConditionalVariable.Broadcast()
	return nil
}
