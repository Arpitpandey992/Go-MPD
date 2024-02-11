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
	go func() {
		<-time.After(time.Second * 10)
		fmt.Println("pausing playback")
		audioPlayer.Pause()
		<-time.After(time.Second * 10)
		fmt.Println("resuming playback")
		audioPlayer.Play()
		<-time.After(time.Second * 10)
		fmt.Println("restarting playback")
		err := audioPlayer.Seek(time.Second * 0)
		if err != nil {
			t.Error(err)
		}
		<-time.After(time.Second * 10)
		fmt.Println("seeking to 23 seconds")
		err = audioPlayer.Seek(time.Second * 23)
		if err != nil {
			t.Error(err)
		}
	}()
	if err != nil {
		t.Error(err)
	}
	for {
		select {
		case <-done:
			return
		case <-time.After(time.Second):
			fmt.Println(audioPlayer.getCurrentPosition())
		}
	}
}
