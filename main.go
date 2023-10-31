package main

import (
	"io"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	// Echo all incoming data.
	io.Copy(conn, conn)
	// Shut down the connection.
	conn.Close()
}

func main() {
	l, err := net.Listen("tcp", ":6969")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConnection(conn)
	}
}
