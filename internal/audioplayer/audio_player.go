package audioplayer

import (
	"log"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
)

type AudioPlayer struct {
	//TODO: Move from StreamSeekCloser to StreamSeeker
	//      Remove some unnecessary Speaker.lock() statements
	Streamer beep.StreamSeekCloser // Streamer is available to manipulate seek position
	Format   beep.Format           // track metadata
	Ctrl     *beep.Ctrl            // for play/pause functionality
}

func (ap *AudioPlayer) Play() {
	if !ap.IsPaused() {
		log.Print("audioplayer is already playing")
		return
	}
	log.Print("debug: locked speaker")
	speaker.Lock()
	defer speaker.Unlock()
	ap.Ctrl.Paused = false
	log.Print("debug: unlocked speaker")
}

func (ap *AudioPlayer) Pause() {
	if ap.IsPaused() {
		log.Print("audioplayer is already paused")
		return
	}
	log.Print("debug: locked speaker")
	speaker.Lock()
	defer speaker.Unlock()
	ap.Ctrl.Paused = true
	log.Print("debug: unlocked speaker")
}

func (ap *AudioPlayer) GetCurrentPosition() time.Duration {
	// TODO: check why Speaker.lock() is recommended here, this might result in race condition
	duration := ap.Format.SampleRate.D(ap.Streamer.Position())
	return duration.Truncate(time.Second)
}

func (ap *AudioPlayer) Seek(seekTime time.Duration) error {
	// Seeks to the given time (granular upto seconds of accuracy)
	log.Printf("seeking to %v", seekTime)
	samplesToSeek := int64(ap.Format.SampleRate) * int64(seekTime.Seconds())

	returnValue := ap.runWithAudioPlayerLock(func() interface{} { return ap.Streamer.Seek(int(samplesToSeek)) })
	if err, ok := returnValue.(error); ok {
		if err != nil {
			log.Printf("error while seeking: %v", err)
		}
		return err
	}
	return nil
}

func (ap *AudioPlayer) IsPaused() bool {
	return ap.Ctrl.Paused
}

func (ap *AudioPlayer) Close() error {
	if !ap.IsPaused() {
		ap.Pause()
	}
	ap.Ctrl = nil
	return ap.Streamer.Close()
}
func (ap *AudioPlayer) runWithAudioPlayerLock(callable func() interface{}) any {
	initiallyPlaying := !ap.IsPaused()
	if initiallyPlaying {
		ap.Pause()
	}
	returnValue := callable()
	if initiallyPlaying {
		ap.Play()
	}
	return returnValue
}
