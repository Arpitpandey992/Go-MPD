package main

import (
	"github.com/arpitpandey992/go-mpd/internal/audioplayer"
	"github.com/arpitpandey992/go-mpd/internal/server"
	"log"
)

func main() {
	musicFile := "music/ricor.flac"
	audioPlayer, err := audioplayer.CreateAudioPlayer(musicFile)
	if err != nil {
		log.Fatal(err)
	}
	defer audioPlayer.Close()
	err = audioPlayer.Play()
	if err != nil {
		log.Fatal(err)
	}
	select {}
	server.StartAndHandleServer()
}
