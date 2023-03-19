package main

import (
	"encoding/json"
	"fmt"
	"net"

	gonanoid "github.com/matoous/go-nanoid"
)

type user struct {
	conn     net.Conn
	username string
}

type Message struct {
	Sender  string
	Request string
	Content string
}

type GameSession struct {
	RoomID  string
	Players []string
	Host    string
}

var users map[string]user
var activeSessions map[string]GameSession

func main() {
	users = make(map[string]user)
	activeSessions = make(map[string]GameSession)
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()

	fmt.Println("Server listening on localhost:8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(conn)

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// defer conn.Close() //need opened connection for later actions

	mess := make(chan Message)
	defer close(mess)

	for {
		go receiveMess(mess, conn)

		t := <-mess
		switch t.Request {
		case "register":
			fmt.Println("register")
			id, err := gonanoid.ID(6)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(id)

			tempUser := user{conn, t.Content}
			fmt.Println(tempUser)

			users[id] = tempUser
			fmt.Println(users)

			data := []byte(id)
			_, err = conn.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		case "createRoom":
			fmt.Println("create room")
			roomID, err := gonanoid.ID(6)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(roomID)
			session := GameSession{roomID, []string{}, t.Sender}

			session.addUser(t.Sender)

			data, err := json.Marshal(session)
			if err != nil {
				fmt.Println(err)
			}

			_, err = users[t.Sender].conn.Write(data)
			if err != nil {
				fmt.Println(err)
			}

			activeSessions[roomID] = session

		case "joinRoom":
			fmt.Println("join room")
			session, ok := activeSessions[t.Content]
			if !ok {
				//session does not exist
				//
			}

			//need to make so 3rd player cant join
			session.Players = append(session.Players, t.Sender)

			message := Message{"server", "gameJoin", users[t.Sender].username}
			data, err := json.Marshal(message)
			if err != nil {
				fmt.Println(err)
			}
			_, err = users[session.Host].conn.Write(data)
			if err != nil {
				fmt.Println(err)
			}

			session.Players[0] = users[session.Players[0]].username
			session.Players[1] = users[session.Players[1]].username
			data, err = json.Marshal(session)
			if err != nil {
				fmt.Println(err)
			}
			_, err = users[t.Sender].conn.Write(data)
			if err != nil {
				fmt.Println(err)
			}

		default:
			fmt.Println("def")
		}
	}
}

func (s *GameSession) addUser(userID string) {
	s.Players = append(s.Players, userID)
}

func receiveMess(c chan Message, con net.Conn) {
	data := make([]byte, 1024)
	n, err := con.Read(data)
	if err != nil {
		fmt.Println(err)
	}
	data = data[:n]

	var mess Message

	err = json.Unmarshal(data, &mess)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(mess)

	c <- mess
}
