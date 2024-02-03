package server

import (
	"io"
	"log"
	"net"
	"strings"
)

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
		_sendWelcomeMessageToConnectionClient(conn)
		go handleConnection(conn)
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

func handleConnection(conn net.Conn) {
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
		err = handleIncomingRequest(string(buf[:n]))
		if err != nil {
			log.Print("error: ", err)
		}
	}
}

func handleIncomingRequest(command string) error {
	chunks := strings.Split(command, " ")
	for i, chunk := range chunks {
		chunks[i] = strings.TrimSpace(chunk)
	}
	requestType := chunks[0]
	if requestType == "audio" {
		err := HandleAudioRequest(chunks[1:])
		if err != nil {
			return err
		}
	}
	return nil
}

func _sendWelcomeMessageToConnectionClient(conn net.Conn) {
	welcomeMessage := "Welcome to Go-MPD!\n"
	_, err := conn.Write([]byte(welcomeMessage))
	if err != nil {
		log.Printf("error: could not send welcome message to %s, error: %s", conn.RemoteAddr(), err)
	}
}
