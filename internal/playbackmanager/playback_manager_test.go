package playbackmanager

import (
	"testing"
)

func TestPlaybackManager(t *testing.T) {
	musicFiles := []string{
		"../../music/sample-3s.mp3",
		"../../music/sample-9s.mp3",
		"../../music/sample-12s.mp3",
	}
	playbackManager := CreatePlaybackManager()
	err := playbackManager.AddAudioFilesToQueue(musicFiles...)
	if err != nil {
		t.Error(err)
	}
	_ = playbackManager.Play()
	<-playbackManager.QueuePlaybackFinished
}
