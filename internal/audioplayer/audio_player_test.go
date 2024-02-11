package audioplayer

import (
	"fmt"
	"testing"
	"time"

	"github.com/gopxl/beep/speaker"
)

func TestAudioPlayer(t *testing.T) {
	musicFile := "../../music/ricor.flac"
	done := make(chan bool)
	audioPlayer, err := CreateAudioPlayer(musicFile, func() { done <- true })
	_ = speaker.Init(audioPlayer.Format.SampleRate, audioPlayer.Format.SampleRate.N(time.Second/10))
	speaker.Play(audioPlayer.Ctrl)
	if err != nil {
		t.Error(err)
	}
	defer audioPlayer.Close()
	audioPlayer.Play()
	if err != nil {
		t.Error(err)
	}
	for {
		select {
		case <-done:
			return
		case <-time.After(time.Second):
			fmt.Println(audioPlayer.getCurrentPositionSeconds())
		}
	}
}
