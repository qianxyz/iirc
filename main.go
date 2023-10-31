package main

import (
	"log"
	"net"
)

// FIXME: Lock the global
var conns = make(map[net.Conn]bool)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	conns[conn] = true
	defer delete(conns, conn)

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if n == 0 && err != nil {
			break
		}

		for c := range conns {
			if c != conn {
				c.Write(buf[:n])
			}
		}
	}
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
