package playbackmanager

import (
	"testing"
	"time"
)

func TestPlaybackManager(t *testing.T) {
	musicFile := "../../music/test.mp3"
	playbackManager := CreatePlaybackManager()
	err := playbackManager.AddAudioToQueue(musicFile)
	if err != nil {
		t.Error(err)
	}
	err = playbackManager.Play()
	if err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Second)
}
