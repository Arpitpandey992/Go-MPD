package playbackmanager

import (
	"testing"
	"time"
)

func TestPlaybackManager(t *testing.T) {
	musicFiles := []string{
		// "../../music/ricor.flac",
		"../../music/sample-96kHz24bit.flac",
		"../../music/sample-3s.mp3",
		"../../music/sample-9s.mp3",
		"../../music/sample-12s.mp3",
	}
	playbackManager := CreatePlaybackManager()
	err := playbackManager.AddAudioFilesToQueue(musicFiles...)
	if err != nil {
		t.Error(err)
	}
	go func() {
		_ = playbackManager.Play()
	}()
	testStop := false
	if testStop {
		time.Sleep(time.Second * 9)
		println("stopping playback of current track")
		_ = playbackManager.Stop()
		time.Sleep(time.Second * 3)
		println("playing again")
		_ = playbackManager.Play()
		// println("testing next()")
		// _ = playbackManager.Next()
	}
	<-playbackManager.QueuePlaybackFinished
}
