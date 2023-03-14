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
	Request string
	Content string
}

var users map[string]user

func main() {
	users = make(map[string]user)
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
	fmt.Println(err)

	fmt.Println(mess)

	c <- mess
}
