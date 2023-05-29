package main

import (
	"encoding/json"
	"fmt"
	"net"

	gonanoid "github.com/matoous/go-nanoid"
)

const (
	REGISTER     = "register"
	CREATEROOM   = "createRoom"
	JOINROOM     = "joinRoom"
	LEAVESESSION = "leaveSession"
	DISCONNECT   = "disconnect"
	STATUS       = "status"
	WIN          = "win"
	REMATCH      = "rematch"
)

type user struct {
	conn     net.Conn
	username string
}

type Message struct {
	Sender  string
	Request string
	// Content []string
	Content map[string]interface{}
	Session GameSession
}

func (m Message) send(conn net.Conn) {
	data, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err)
	}
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println(err)
	}
}

type GameSession struct {
	RoomID  string
	Players []string
	Host    string
	Turn    string
}

func (s *GameSession) addUser(userID string) bool {
	//TODO: prevent having more than 2 players in a session
	if len(s.Players) < 2 {
		s.Players = append(s.Players, userID)
		return true
	}
	return false
}

func (s *GameSession) deleteUser(userID string) bool {
	fmt.Println(s.Players)
	fmt.Println(userID, " wants to leave")
	for i, v := range s.Players {
		if v == userID {
			if i == 0 {
				s.Players = s.Players[i:]
				fmt.Println("i is 0 ", s.Players)

				return true
			} else {
				s.Players = s.Players[:i]
				fmt.Println("i is 1 ", s.Players)

				return true
			}
		}
	}

	fmt.Println("outside loop ", s.Players)

	return false
}

var users map[string]user
var activeSessions map[string]*GameSession

func main() {
	users = make(map[string]user)
	activeSessions = make(map[string]*GameSession)
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
		case REGISTER:
			fmt.Println("register")
			id, err := gonanoid.ID(6)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(id)

			tempUser := user{conn, t.Content["username"].(string)}
			fmt.Println(tempUser)

			users[id] = tempUser
			fmt.Println(users)

			data := []byte(id)
			_, err = conn.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		case CREATEROOM:
			fmt.Println("create room")
			roomID, err := gonanoid.ID(6)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(roomID)
			session := GameSession{roomID, []string{}, t.Sender, t.Sender}

			session.addUser(t.Sender)

			data, err := json.Marshal(session)
			if err != nil {
				fmt.Println(err)
			}

			_, err = users[t.Sender].conn.Write(data)
			if err != nil {
				fmt.Println(err)
			}

			activeSessions[roomID] = &session

			fmt.Println("CURRENT PLAYERS", session.Players)

		case JOINROOM:
			fmt.Println("join room")
			session, ok := activeSessions[t.Content["roomID"].(string)]
			if !ok {
				temp := GameSession{}
				data, err := json.Marshal(temp)
				if err != nil {
					fmt.Println(err)
				}
				_, err = users[t.Sender].conn.Write(data)
				if err != nil {
					fmt.Println(err)
				}
				break
			}

			fmt.Println("CURRENT PLAYERS", session.Players)

			if !session.addUser(t.Sender) {
				temp := GameSession{}
				data, err := json.Marshal(temp)
				if err != nil {
					fmt.Println(err)
				}
				_, err = users[t.Sender].conn.Write(data)
				if err != nil {
					fmt.Println(err)
				}
				break
			}

			fmt.Println("CURRENT PLAYERS AFTER  JOIN", session.Players)

			message := Message{"server", "gameJoin", map[string]interface{}{"username": users[t.Sender].username}, *session}
			message.send(users[session.Host].conn)

			temp := GameSession{
				session.RoomID,
				[]string{users[session.Players[0]].username, users[session.Players[1]].username},
				session.Host,
				session.Host,
			}
			data, err := json.Marshal(temp)
			if err != nil {
				fmt.Println(err)
			}
			_, err = users[t.Sender].conn.Write(data)
			if err != nil {
				fmt.Println(err)
			}

			// message = Message{"server", "sessionUpdate", "", *session}
			// message.send(users[session.Host].conn)

			// data, err = json.Marshal(session)
			// if err != nil {
			// 	fmt.Println(err)
			// }

			// _, err = users[session.Host].conn.Write(data)
			// if err != nil {
			// 	fmt.Println(err)
			// }

			fmt.Println(session.Players)
		case LEAVESESSION:
			fmt.Println("leave")

			session := activeSessions[t.Content["roomID"].(string)]

			fmt.Println("CURRENT PLAYERS: ", session.Players)

			if session.deleteUser(t.Sender) {
				if t.Sender == session.Host {
					fmt.Println("host leave")
					message := Message{"server", "leave", map[string]interface{}{"status": "success"}, *session}
					message.send(users[t.Sender].conn)

					fmt.Println("CURRENT PLAYERS: ", session.Players, ", HOST: ", session.Host)

					delete(activeSessions, session.RoomID)

					fmt.Println("ACTIVE SESSIONS: ", activeSessions)
					if len(session.Players) > 1 {
						message = Message{"server", "sessionDisbanded", map[string]interface{}{}, *session}
						message.send(users[session.Players[1]].conn)
					}

				} else {
					fmt.Println("non host leave")

					fmt.Println("CURRENT PLAYERS: ", session.Players, ", HOST: ", session.Host)

					message := Message{"server", "leave", map[string]interface{}{"status": "success"}, *session}
					message.send(users[t.Sender].conn)

					message = Message{"server", "sessionLeave", map[string]interface{}{}, *session} //
					message.send(users[session.Host].conn)
				}
			} else {

				fmt.Println("CURRENT PLAYERS: ", session.Players, ", HOST: ", session.Host)

				message := Message{"server", "leave", map[string]interface{}{"status": "success"}, *session}
				message.send(users[t.Sender].conn)
			}
		case DISCONNECT:
			fmt.Println(users)
			users[t.Sender].conn.Close()
			delete(users, t.Sender)
			fmt.Println(users)
			return
		case STATUS:
			session := activeSessions[t.Content["roomID"].(string)]

			_, gameEnd := t.Content["gameEnd"]
			if gameEnd {
				_, rematch := t.Content["rematch"]
				if rematch {
					// //if both players want to rematch send message to start new game
				} else {
					message := Message{"server", t.Request, t.Content, t.Session}
					message.send(users[t.Session.Players[1]].conn)
					message.send(users[t.Session.Players[0]].conn)
				}
			} else {
				if t.Sender == session.Turn {
					if session.Turn == session.Players[0] {
						session.Turn = session.Players[1]
					} else {
						session.Turn = session.Players[0]
					}
					message := Message{"server", t.Request, map[string]interface{}{"move": t.Content["move"], "player": t.Content["player"]}, *session}

					message.send(users[session.Players[1]].conn)
					message.send(users[session.Players[0]].conn)
				}
			}
		default:
			fmt.Println("def")
		}
	}
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
