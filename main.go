package main

import (
	"fmt"
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

	client := &client{conn: conn, nick: "Anon"}

	clients[client] = true
	defer delete(clients, client)

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if n == 0 && err != nil {
			break
		}

		if buf[0] != '/' {
			for c := range clients {
				if c == client {
					continue
				}
				fmt.Fprintf(c.conn, "%s: %s", client.nick, buf[:n])
			}
			continue
		}

		fields := strings.Fields(string(buf[:n]))
		switch fields[0] {
		case "/nick":
			if len(fields) != 2 {
				fmt.Fprintln(client.conn, "/nick: bad arguments")
				continue
			}
			client.nick = fields[1]
			fmt.Fprintf(client.conn, "Nickname changed to %s\n", client.nick)
		default:
			fmt.Fprintf(client.conn, "Unknown command: %s\n", fields[0])
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
