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

func connectToServer() bool {
	connection, err := net.Dial("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return false
	}
	// defer connection.Close() //need opened connection for later actions

	message := struct {
		Request string
		Content string
	}{"register", username}

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
	gameWindow := myApp.NewWindow("Tic-Tac-Toe")
	gameWindow.Resize(fyne.NewSize(600, 600))

	btnJoin := widget.NewButton("Join Room", func() {

	})
	btnCreate := widget.NewButton("Create Room", func() {
		createRoom()
	})

	content := container.NewCenter(container.NewVBox(btnJoin, btnCreate))
	gameWindow.SetContent(content)

	gameWindow.Show()
}

func createRoom() {

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
