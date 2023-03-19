package main

import (
	"encoding/json"
	"fmt"
	"net"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

//TODO: username max lenght = 15 //data := make([]byte, 1024) -> data := make([]byte, 15) //

var myApp fyne.App
var userID string
var username string
var connection net.Conn
var session Session

type Session struct {
	RoomID  string
	Players []string
	Host    string
}

type Message struct {
	Sender  string
	Request string
	Content string
}

func connectToServer() bool {
	message := Message{userID, "register", username}

	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println(err)
		return false
	}

	connection, err = net.Dial("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return false
	}
	// defer connection.Close() //need opened connection for later actions

	_, err = connection.Write(data)
	if err != nil {
		fmt.Println(err)
		return false
	}

	data = make([]byte, 6)
	_, err = connection.Read(data)
	if err != nil {
		fmt.Println(err)
	}

	userID = string(data)

	fmt.Println(userID)
	fmt.Println(username)
	return true
}

func initializeGameWindow() {
	window := myApp.NewWindow("Tic-Tac-Toe")
	window.Resize(fyne.NewSize(600, 600))

	btnJoin := widget.NewButton("Join Room", func() {
		temp := myApp.NewWindow("Join room")
		temp.Resize(fyne.NewSize(600, 600))

		roomID := widget.NewEntry()
		content := container.NewCenter(container.NewVBox(roomID, widget.NewButton("Join", func() {
			if joinRoom(roomID.Text) {
				temp.Close()
				label1 := widget.NewLabel(session.RoomID)
				label2 := widget.NewLabel(session.Players[0])
				label3 := widget.NewLabel(session.Players[1])
				content := container.NewVBox(label1, label2, label3)
				window.SetContent(container.NewCenter(content))
				// messageChan := make(chan Message)
				// go receiveMess(messageChan)
				// go gameStatusUpdates(messageChan, content)
			}
		})))
		temp.SetContent(content)

		temp.Show()
	})
	btnCreate := widget.NewButton("Create Room", func() {
		if createGameRoom() {
			label1 := widget.NewLabel(session.RoomID)
			label2 := widget.NewLabel(username)
			content := container.NewVBox(label1, label2)
			window.SetContent(container.NewCenter(content))
			messageChan := make(chan Message)
			go receiveMess(messageChan)
			go gameStatusUpdates(messageChan, content)
		}
	})

	content := container.NewCenter(container.NewVBox(btnJoin, btnCreate))
	window.SetContent(content)

	window.Show()
}

func gameStatusUpdates(c chan Message, container *fyne.Container) {
	//listen for game changes
	//user leaves
	//game progress
	t := <-c
	switch t.Request {
	case "gameJoin":
		fmt.Println("game join")
		container.Add(widget.NewLabel(t.Content))
		container.Refresh()
	default:
		fmt.Println("def")
	}
}

func receiveMess(c chan Message) {
	data := make([]byte, 1024)
	n, err := connection.Read(data)
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

func joinRoom(roomID string) bool {
	message := Message{userID, "joinRoom", roomID}

	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println(err)
		return false
	}
	_, err = connection.Write(data)
	if err != nil {
		fmt.Println(err)
		return false
	}

	data = make([]byte, 1024)
	n, err := connection.Read(data)
	if err != nil {
		fmt.Println(err)
		return false
	}
	data = data[:n]

	err = json.Unmarshal(data, &session)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func createGameRoom() bool {
	fmt.Println(connection)

	message := Message{userID, "createRoom", ""}

	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println(err)
		return false
	}
	_, err = connection.Write(data)
	if err != nil {
		fmt.Println(err)
		return false
	}

	//server sents back roomID, current players

	data = make([]byte, 1024)
	n, err := connection.Read(data)
	if err != nil {
		fmt.Println(err)
		return false
	}
	data = data[:n]

	err = json.Unmarshal(data, &session)
	if err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Println(session)
	return true
}

func main() {
	myApp = app.New()

	loginWindow := myApp.NewWindow("Tic-Tac-Toe")
	loginWindow.Resize(fyne.NewSize(600, 600))

	usernameEntry := widget.NewEntry()
	form := widget.NewForm(widget.NewFormItem("Username:", usernameEntry),
		widget.NewFormItem("", widget.NewButton("Connect", func() {
			username = usernameEntry.Text
			if connectToServer() {
				loginWindow.Close()
				initializeGameWindow()
			}
		})))
	form.Resize(fyne.NewSize(200, 200))
	form.Move(fyne.NewPos(150, 200))

	content := container.NewWithoutLayout(form)
	loginWindow.SetContent(content)

	loginWindow.Show()
	myApp.Run()
}
