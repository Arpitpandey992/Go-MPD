package playbackmanager

import (
	"fmt"
	"log"
	"time"

	"github.com/arpitpandey992/go-mpd/internal/audioplayer"
	"github.com/gopxl/beep/speaker"
)

type PlaybackManager struct {
	playbackQueue         []string
	queuePosition         int
	AudioPlayer           *audioplayer.AudioPlayer
	playbackFinished      chan bool
	newTrackAdded         chan bool
	newAudioPlayerCreated chan bool
}

func CreatePlaybackManager() *PlaybackManager {
	playbackManager := PlaybackManager{
		playbackQueue:         []string{},
		queuePosition:         0,
		AudioPlayer:           nil,
		playbackFinished:      make(chan bool),
		newTrackAdded:         make(chan bool),
		newAudioPlayerCreated: make(chan bool),
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
	pm.newTrackAdded <- true //signal the channel to proceed with the creation of AudioPlayer for the new track if it is waiting
	return nil
}

func (pm *PlaybackManager) Play() error {
	if pm.queuePosition >= len(pm.playbackQueue) {
		return fmt.Errorf("reached the end of playback queue")
	}
	if pm.AudioPlayer == nil {
		// waiting for the audioplayer of the current track to be created
		<-pm.newAudioPlayerCreated
	}
	if !pm.AudioPlayer.IsPaused() {
		return fmt.Errorf("%s is already playing", pm.playbackQueue[pm.queuePosition])
	}

	pm.AudioPlayer.Play()
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
		<-pm.newTrackAdded // creation of new AudioPlayer is blocked till a new track is added
		log.Printf("creating audioplayer for %s", pm.playbackQueue[pm.queuePosition])
		err := pm.createAudioPlayerForCurrentTrack()
		if err != nil {
			log.Printf("error while creating audio player for %s [skipped], error: %v", pm.playbackQueue[pm.queuePosition], err)
		} else {
			<-pm.playbackFinished // blocking till playback is finished
			err = pm.doOnTrackPlaybackFinished()
			if err != nil {
				log.Print(err)
			}
		}
	}
}

func (pm *PlaybackManager) createAudioPlayerForCurrentTrack() error {
	doOnFinishPlaying := func() {
		pm.playbackFinished <- true
	}
	ap, err := audioplayer.CreateAudioPlayer(pm.playbackQueue[pm.queuePosition], doOnFinishPlaying)
	if err != nil {
		return err
	}
	pm.AudioPlayer = ap
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
	err := speaker.Init(pm.AudioPlayer.Format.SampleRate, pm.AudioPlayer.Format.SampleRate.N(time.Second/10))
	if err != nil {
		return err
	}
	speaker.Play(pm.AudioPlayer.Ctrl)
	return nil
}

func (pm *PlaybackManager) doOnTrackPlaybackFinished() error {
	log.Printf("%s finished playing", pm.playbackQueue[pm.queuePosition])
	pm.AudioPlayer.Close()
	pm.AudioPlayer = nil
	pm.queuePosition++
	return pm.Play() // if there is a new track available, play it
}
