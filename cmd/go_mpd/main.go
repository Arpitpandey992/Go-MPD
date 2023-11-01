package main

import (
	"log"

	"github.com/arpitpandey992/go-mpd/internal/audioplayer"
)

func main() {
	musicFile := "music/ricor.mp3"
	audioPlayer, err := audioplayer.CreateAudioPlayer(musicFile)
	if err != nil {
		log.Fatal(err)
	}
	err = audioPlayer.Play()
	if err != nil {
		log.Fatal(err)
	}
}
