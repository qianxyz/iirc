package main

import (
	"log"
	"net"
	"strings"
)

type client struct {
	conn net.Conn
	nick string
}

// FIXME: Lock the global
var clients = make(map[*client]bool)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	client := client{conn: conn, nick: "Anon"}

	clients[&client] = true
	defer delete(clients, &client)

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if n == 0 && err != nil {
			break
		}

		if buf[0] != '/' {
			for c := range clients {
				if c == &client {
					continue
				}
				c.conn.Write([]byte(client.nick))
				c.conn.Write([]byte(": "))
				c.conn.Write(buf[:n])
			}
			continue
		}

		fields := strings.Fields(string(buf[:n]))
		switch fields[0] {
		case "/nick":
			// FIXME: Check # of args
			client.nick = fields[1]
			client.conn.Write([]byte("Nickname changed to "))
			client.conn.Write([]byte(client.nick))
			client.conn.Write([]byte("\n"))
		default:
			client.conn.Write([]byte("Unknown command: "))
			client.conn.Write([]byte(fields[0]))
			client.conn.Write([]byte("\n"))
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
