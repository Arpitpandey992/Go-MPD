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
	speaker.Lock()
	ap.Ctrl.Paused = false
	speaker.Unlock()
}

func (ap *AudioPlayer) Pause() {
	if ap.IsPaused() {
		log.Print("audioplayer is already paused")
		return
	}
	speaker.Lock()
	ap.Ctrl.Paused = true
	speaker.Unlock()
}

func (ap *AudioPlayer) getCurrentPosition() time.Duration {
	// TODO: check why Speaker.lock() is recommended here, this might result in race condition
	duration := ap.Format.SampleRate.D(ap.Streamer.Position())
	return duration.Truncate(time.Second)
}

func (ap *AudioPlayer) Seek(seekTime time.Duration) error {
	// Seeks to the given time (granular upto seconds of accuracy)
	log.Printf("seeking to %v", seekTime)
	samplesToSeek := int64(ap.Format.SampleRate) * int64(seekTime.Seconds())
	ap.Pause()
	err := ap.Streamer.Seek(int(samplesToSeek))
	ap.Play()
	if err != nil {
		log.Printf("error while seeking: %v", err)
	}
	return err
}

func (ap *AudioPlayer) IsPaused() bool {
	return ap.Ctrl.Paused
}

func (ap *AudioPlayer) Close() error {
	return ap.Streamer.Close()
}
