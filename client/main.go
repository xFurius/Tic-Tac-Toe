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

	})
	btnCreate := widget.NewButton("Create Room", func() {
		if createGameRoom() {
			window.SetContent(widget.NewButton("GAME SESSION CREATED", nil))
		}
	})

	content := container.NewCenter(container.NewVBox(btnJoin, btnCreate))
	window.SetContent(content)

	window.Show()
}

func createGameRoom() bool {
	//send request to server

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

	session := struct {
		RoomID  string
		Players []string
		Host    string
	}{}

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
