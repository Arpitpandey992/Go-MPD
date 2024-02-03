package audioplayer

import (
	"testing"
	"time"
)

func TestAudioPlayer(t *testing.T) {
	musicFile := "../../music/ricor.flac"
	audioPlayer, err := CreateAudioPlayer(musicFile)
	if err != nil {
		t.Error(err)
	}
	defer audioPlayer.Close()
	err = audioPlayer.Play()
	if err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Second)
}
