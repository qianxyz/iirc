package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type Client struct {
	conn net.Conn
	nick string
	room *Room
}

type Room struct {
	name    string
	clients map[*Client]bool
}

// FIXME: Lock the global
var (
	clients = make(map[*Client]bool)
	rooms   = make(map[string]*Room)
)

func handleConnection(conn net.Conn) {
	client := &Client{conn: conn, nick: "Anon"}
	clients[client] = true

	buf := make([]byte, 1024)
outer:
	for {
		n, err := conn.Read(buf)
		if n == 0 && err != nil {
			break
		}

		if buf[0] != '/' {
			if client.room == nil {
				// TODO: tell client
				continue
			}
			for c := range client.room.clients {
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
		case "/join":
			if len(fields) != 2 {
				fmt.Fprintln(client.conn, "/join: bad arguments")
				continue
			}

			// find the room, create if not exist
			name := fields[1]
			room, ok := rooms[name]
			if !ok {
				room = &Room{name: name, clients: make(map[*Client]bool)}
				rooms[name] = room
			}

			// register the client with the room
			room.clients[client] = true
			client.room = room

			// TODO: notify room members
		case "/leave":
			// TODO: notify room members
			if client.room != nil {
				delete(client.room.clients, client)
				if len(client.room.clients) == 0 {
					delete(rooms, client.room.name)
				}
				client.room = nil
			}
		case "/quit":
			break outer
		default:
			fmt.Fprintf(client.conn, "Unknown command: %s\n", fields[0])
		}
	}

	delete(clients, client)
	if client.room != nil {
		delete(client.room.clients, client)
		if len(client.room.clients) == 0 {
			delete(rooms, client.room.name)
		}
	}
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
