package audioplayer

import (
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"log"
)

type AudioPlayer struct {
	streamer beep.StreamSeekCloser
	Format   beep.Format
	Ctrl     *beep.Ctrl
}

func (ap *AudioPlayer) Play() {
	if !ap.IsPaused() {
		log.Print("audioplayer is already playing")
	}
	speaker.Lock()
	ap.Ctrl.Paused = false
	speaker.Unlock()
}

func (ap *AudioPlayer) Pause() {
	if ap.IsPaused() {
		log.Print("audioplayer is already paused")
	}
	speaker.Lock()
	ap.Ctrl.Paused = true
	speaker.Unlock()
}

func (ap *AudioPlayer) IsPaused() bool {
	return ap.Ctrl.Paused
}

func (ap *AudioPlayer) Close() error {
	return ap.streamer.Close()
}
