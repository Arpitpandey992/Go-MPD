package playbackmanager

import (
	"testing"
	"time"
)

func TestPlayPauseStop(t *testing.T) {
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
	_ = playbackManager.Play()
	testStop := true
	if testStop {
		time.Sleep(time.Second * 9)
		println("stopping playback of current track")
		_ = playbackManager.Stop()
		time.Sleep(time.Second * 3)
		println("playing again")
		_ = playbackManager.Play()
		time.Sleep(time.Second * 4)
		println("pausing for 3 seconds")
		_ = playbackManager.Pause()
		time.Sleep(time.Second * 3)
		println("playing")
		_ = playbackManager.Play()
	}
	<-playbackManager.QueuePlaybackFinished
}
func TestAutoNext(t *testing.T) {
	musicFiles := []string{
		"../../music/sample-3s.mp3",
		"../../music/sample-9s.mp3",
		"../../music/sample-12s.mp3",
		"../../music/sample-96kHz24bit.flac",
	}
	playbackManager := CreatePlaybackManager()
	err := playbackManager.AddAudioFilesToQueue(musicFiles...)
	if err != nil {
		t.Error(err)
	}
	_ = playbackManager.Play()
	<-playbackManager.QueuePlaybackFinished
}
func TestNextPrevious(t *testing.T) {
	musicFiles := []string{
		"../../music/ricor.flac",
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
	time.Sleep(time.Second * 5)
	println("Moving to next track")
	_ = playbackManager.Next()
	time.Sleep(time.Second * 4)
	println("Moving to previous track")
	_ = playbackManager.Previous()
	time.Sleep(time.Second * 6)
	println("skipping 2 tracks")
	_ = playbackManager.Next()
	_ = playbackManager.Next()
	<-playbackManager.QueuePlaybackFinished
}
