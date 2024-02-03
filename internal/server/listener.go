package server

import (
	"io"
	"log"
	"net"
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
	buf := make([]byte, 50)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Print(err)
			}
			return
		}
		log.Printf("received: %q", buf[:n])
	}
}
