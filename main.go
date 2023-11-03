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
var rooms = make(map[string]*Room)

func (c *Client) join(room *Room) {
	if c.room != nil {
		c.leave()
	}

	room.clients[c] = true
	c.room = room

	msg := fmt.Sprintf("%s joined the room.\n", c.nick)
	room.broadcast(msg, c)
}

func (c *Client) leave() {
	room := c.room
	if room == nil {
		return
	}

	msg := fmt.Sprintf("%s left the room.\n", c.nick)
	room.broadcast(msg, c)

	delete(room.clients, c)
	if len(room.clients) == 0 {
		delete(rooms, room.name)
	}
	c.room = nil
}

func (r *Room) broadcast(msg string, exclude *Client) {
	for c := range r.clients {
		if c == exclude {
			continue
		}
		fmt.Fprint(c.conn, msg)
	}
}

func handleConnection(conn net.Conn) {
	client := &Client{conn: conn, nick: "Anon"}

	buf := make([]byte, 1024)
outer:
	for {
		n, err := conn.Read(buf)
		if n == 0 && err != nil {
			break
		}

		if buf[0] != '/' {
			if client.room == nil {
				fmt.Fprintln(client.conn, "You are not in a room.")
				fmt.Fprintln(client.conn, "Use `/join` to join one.")
			} else {
				msg := fmt.Sprintf("%s: %s", client.nick, buf[:n])
				client.room.broadcast(msg, client)
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
			fmt.Fprintf(client.conn, "Nickname changed to %s.\n", client.nick)
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

			client.join(room)
		case "/leave":
			if client.room == nil {
				fmt.Fprintln(client.conn, "You are not in a room.")
			} else {
				client.leave()
			}
		case "/quit":
			break outer
		default:
			fmt.Fprintf(client.conn, "Unknown command: %s\n", fields[0])
		}
	}

	client.leave()
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
