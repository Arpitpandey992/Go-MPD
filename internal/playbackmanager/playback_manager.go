package playbackmanager

import (
	"fmt"
	"log"
	"time"

	"github.com/arpitpandey992/go-mpd/internal/audioplayer"
	"github.com/gopxl/beep/speaker"
)

type PlaybackManager struct {
	playbackQueue    []string
	queuePosition    int
	playbackFinished chan bool
	AudioPlayer      *audioplayer.AudioPlayer
}

func GetNewPlaybackManager() *PlaybackManager {
	playbackManager := PlaybackManager{
		playbackQueue:    []string{},
		queuePosition:    0,
		playbackFinished: make(chan bool),
		AudioPlayer:      nil,
	}
	go playbackManager.waitAndManagePlayback()
	return &playbackManager
}

func (pm *PlaybackManager) AddAudioToQueue(filePath string) error {
	err := audioplayer.IsFileSupported(filePath)
	if err != nil {
		log.Print("error: ", err)
		return err
	}
	pm.playbackQueue = append(pm.playbackQueue, filePath)
	return nil
}

func (pm *PlaybackManager) Play() error {
	if pm.queuePosition >= len(pm.playbackQueue) {
		return fmt.Errorf("nothing left in queue to play")
	}
	if pm.AudioPlayer != nil && !pm.AudioPlayer.IsPaused() {
		return fmt.Errorf("%s is already playing", pm.playbackQueue[pm.queuePosition])
	}
	doOnFinishPlaying := func() {
		pm.playbackFinished <- true
	}
	ap, err := audioplayer.CreateAudioPlayer(pm.playbackQueue[pm.queuePosition], doOnFinishPlaying)
	pm.AudioPlayer = ap
	if err != nil {
		return err
	}
	err = pm.initSpeakerOrResample()
	if err != nil {
		return err
	}
	ap.Play()
	return nil
}

func (pm *PlaybackManager) Pause() error {
	if pm.queuePosition >= len(pm.playbackQueue) || pm.AudioPlayer == nil || pm.AudioPlayer.IsPaused() {
		return fmt.Errorf("no song is playing")
	}
	pm.AudioPlayer.Pause()
	return nil
}

func (pm *PlaybackManager) waitAndManagePlayback() {
	for {
		<-pm.playbackFinished //blocking on the playback being finished
		log.Printf("%s finished playing", pm.playbackQueue[pm.queuePosition])
		pm.queuePosition++
		err := pm.Play()
		if err != nil {
			log.Print(err)
		}
	}
}

func (pm *PlaybackManager) initSpeakerOrResample() error {
	// make sure this init call only happens if the new samplerate is different from the earlier playing samplerate
	// resample should be used here, but there is a quality concern as it resamples it
	err := speaker.Init(pm.AudioPlayer.Format.SampleRate, pm.AudioPlayer.Format.SampleRate.N(time.Second/10))
	if err != nil {
		return err
	}
	speaker.Play(pm.AudioPlayer.Ctrl)
	return nil
}
