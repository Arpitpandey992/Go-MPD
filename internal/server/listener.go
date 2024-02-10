package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type Handlers struct {
	audioRequestHandler *AudioRequestsHandler
}

func StartAndHandleServer() {
	SERVER_PROTOCOL := "tcp"
	SERVER_ADDRESS := "127.0.0.1:6600"
	listener := getListener(SERVER_PROTOCOL, SERVER_ADDRESS)
	defer listener.Close()
	handleIncomingConnections(listener)
}

func handleIncomingConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept incoming connection request, error: %v", err)
			continue
		}
		log.Print("successfully connected with incoming client")
		handlers := &Handlers{audioRequestHandler: getNewAudioRequestsHandler()}
		sendWelcomeMessageToConnectionClient(conn)
		go handleConnection(conn, handlers)
	}
}

func getListener(protocol string, server_address string) net.Listener {
	listener, err := net.Listen(protocol, server_address)
	if err != nil {
		log.Fatalf("cannot start the server at: %s\n%v", server_address, err)
	}
	log.Printf("listening on: %v", listener.Addr().String())
	return listener
}

func handleConnection(conn net.Conn, handlers *Handlers) {
	defer conn.Close()
	buf := make([]byte, 2500)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Print(err)
			}
			return
		}
		log.Printf("received: %q", buf[:n])
		err = handleIncomingRequest(string(buf[:n]), conn, handlers)
		if err != nil {
			log.Print("error: ", err)
			_ = sendMessageToConnectionClient("error: "+err.Error(), conn)
		}
	}
}

func handleIncomingRequest(command string, conn net.Conn, handlers *Handlers) error {
	chunks := breakCommandIntoChunks(command)
	log.Printf("commands: %s", strings.Join(chunks, ", "))
	for i, chunk := range chunks {
		chunks[i] = strings.TrimSpace(chunk)
	}
	if len(chunks) == 0 {
		return nil
	}
	requestType := chunks[0]
	switch requestType {
	case "audio":
		if len(chunks) < 2 {
			return fmt.Errorf("audio command expects at least one argument")
		}
		returnMessage, err := handlers.audioRequestHandler.HandleAudioRequest(chunks[1:])
		if err != nil {
			return err
		}
		if returnMessage != "" {
			_ = sendMessageToConnectionClient(returnMessage, conn)
		}
	default:
		return fmt.Errorf("invalid request type: %s", requestType)
	}
	return nil
}

func sendWelcomeMessageToConnectionClient(conn net.Conn) {
	welcomeMessage := "Welcome to Go-MPD!"
	err := sendMessageToConnectionClient(welcomeMessage, conn)
	if err != nil {
		log.Printf("error: could not send welcome message to %s, error: %s", conn.RemoteAddr(), err)
	}
}

func sendMessageToConnectionClient(message string, conn net.Conn) error {
	_, err := conn.Write([]byte(message + "\n"))
	return err
}

func breakCommandIntoChunks(command string) []string {
	command = strings.TrimSpace(command)
	chunks := []string{}
	i, n := 0, len(command)
	for i < n {
		j := i + 1
		if command[i] == '"' {
			for j < n && command[j] != '"' {
				j++
			}
			chunks = append(chunks, command[i+1:j])
			i = j + 1
		} else {
			for j < n && command[j] != ' ' {
				j++
			}
			chunks = append(chunks, command[i:j])
			i = j + 1
		}
	}
	return chunks
}
