package audioplayer

import (
	"log"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
)

type AudioPlayer struct {
	//TODO: Move from StreamSeekCloser to StreamSeeker
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

func (ap *AudioPlayer) getCurrentPositionSeconds() time.Duration {
	speaker.Lock()
	duration:= ap.Format.SampleRate.D(ap.Streamer.Position()).Round(time.Second)
	speaker.Unlock()
	return duration
}

func (ap *AudioPlayer) IsPaused() bool {
	return ap.Ctrl.Paused
}

func (ap *AudioPlayer) Close() error {
	return ap.Streamer.Close()
}
