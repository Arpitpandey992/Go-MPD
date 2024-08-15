package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arpitpandey992/go-mpd/internal/config"
	"github.com/arpitpandey992/go-mpd/internal/database"
)

var setupComplete bool
var server *Server
var conn net.Conn

func TestMain(m *testing.M) {
	if !setupComplete {
		err := setup()
		if err != nil {
			fmt.Printf("error while setup: %s", err.Error())
			os.Exit(1)
		}
		setupComplete = true
	}

	exitVal := m.Run()

	teardown()

	os.Exit(exitVal)
}

func setup() error {
	var err error
	println("starting server")
	config, err := config.GetBaseConfiguration()
	if err != nil {
		return err
	}
	db := database.GetNewAudioMeiliSearchClient(config)
	server = CreateAndStartServer(db)
	println("connecting to server")
	conn, err = net.Dial("tcp", server.Address)
	if err != nil {
		return err
	}
	welcomeMessage, err := bufio.NewReader(conn).ReadString(server.Delimiter[0]) // ignore the welcome message
	if err != nil {
		return err
	}
	log.Print(welcomeMessage)
	return nil
}

func teardown() {
	println("Stopping Server")
	conn.Close()
	server.Close()
}

func TestServerPing(t *testing.T) {
	message := "ping"
	var err error
	_, err = conn.Write([]byte(message))
	checkError(err, t)
	response := extractLineFromConnection(t)
	expectedResult := "pong" + server.Delimiter
	if response != expectedResult {
		t.Errorf("expected: %s, got: %s", expectedResult, response)
	}
}

func TestServerSongPlayback(t *testing.T) {
	baseMusicPath := "../../music" //TODO: move from these hardcoded paths to using filepath package. This will not work in windows
	musicFiles := []string{
		"ricor.flac",
		"sample-96kHz24bit.flac",
		"sample-3s.mp3",
		"sample-9s.mp3",
		"sample-12s.mp3",
	}
	for _, musicFile := range musicFiles {
		message := "audio add " + filepath.Join(baseMusicPath, musicFile)
		sendMessageToServer(message, t)
		log.Print(extractLineFromConnection(t))
	}

	sendMessageToServer("audio play", t)
	time.Sleep(3 * time.Second)
	sendMessageToServer("audio pause", t)
	time.Sleep(3 * time.Second)
	sendMessageToServer("audio next", t)
	time.Sleep(3 * time.Second)
	sendMessageToServer("audio play", t)
	time.Sleep(3 * time.Second)
	sendMessageToServer("audio stop", t)
	time.Sleep(3 * time.Second)
	sendMessageToServer("audio play", t)
	time.Sleep(3 * time.Second)
	sendMessageToServer("audio next", t)
	time.Sleep(3 * time.Second)
	sendMessageToServer("audio prev", t)
	time.Sleep(3 * time.Second)
	sendMultipleMessagesToServer([]string{"audio next", "audio next"}, t)
	time.Sleep(7 * time.Second)
}

func sendMessageToServer(message string, t *testing.T) {
	_, err := conn.Write([]byte(message))
	checkError(err, t)
}

func extractLineFromConnection(t *testing.T) string {
	response, err := bufio.NewReader(conn).ReadString(server.Delimiter[0])
	checkError(err, t)
	log.Printf("received: %s", response)
	return response
}

func checkError(err error, t *testing.T) {
	if err != nil {
		t.Error(err)
	}
}

func sendMultipleMessagesToServer(messages []string, t *testing.T) {
	sendMessageToServer(strings.Join(messages, server.Delimiter), t)
}
