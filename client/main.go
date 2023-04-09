package main

import (
	"encoding/json"
	"fmt"
	"net"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var myApp fyne.App
var userID string
var username string
var connection net.Conn
var session Session

type Session struct {
	RoomID  string
	Players []string //
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
	var contentMain *fyne.Container
	window := myApp.NewWindow("Tic-Tac-Toe")
	window.Resize(fyne.NewSize(600, 600))
	roomID := widget.NewEntry()

	btnJoin := widget.NewButton("Join Room", func() {
		go func() {
			fmt.Println("num of goroutines: ", runtime.NumGoroutine())

			if joinRoom(roomID.Text) {
				label1 := widget.NewLabel(session.RoomID)
				label2 := widget.NewLabel(session.Players[0])
				label3 := widget.NewLabel(session.Players[1])
				leaveBtn := widget.NewButton("Quit session", func() { //
					go func() {
						if leaveSession() {
							window.SetContent(contentMain)
							fmt.Println("num of goroutines: ", runtime.NumGoroutine())
							return
						}
					}()
					fmt.Println("num of goroutines: ", runtime.NumGoroutine())
				})
				// content := container.NewVBox(leaveBtn, label1, label2, label3)
				// window.SetContent(container.NewCenter(content))
				iconRes, err := fyne.LoadResourceFromPath("./assets/circle.png")
				if err != nil {
					fmt.Println(err)
				}
				content1 := container.NewVBox(container.NewCenter(container.NewHBox(label1, widget.NewIcon(iconRes), label2, widget.NewIcon(iconRes), label3)), container.NewCenter(container.NewHBox(leaveBtn)))
				content2 := container.NewAdaptiveGrid(3, widget.NewButton("X", nil), widget.NewButton("X", nil), widget.NewButton("X", nil), widget.NewButton("X", nil), widget.NewButton("X", nil), widget.NewButton("X", nil), widget.NewButton("X", nil), widget.NewButton("X", nil), widget.NewButton("X", nil))
				window.SetContent(container.NewBorder(content1, nil, nil, nil, content2))

				messageChanUser := make(chan Message)
				for {
					go receiveMess(messageChanUser)

					t := <-messageChanUser
					fmt.Println(t)
					switch t.Request {
					case "gameJoin":
						fmt.Println("game join")
						label3.SetText(t.Content)
						content1.Refresh()
					case "sessionLeave":
						fmt.Println("session Leave")
						label3.SetText("")
						content1.Refresh()
						return
					case "sessionDisbanded":
						fmt.Println("sessionDisband")
					case "leave":
						fmt.Println("leave ", t.Content)
						if t.Content == "success" {
							fmt.Println("success")
							window.SetContent(contentMain)
						}
						return
					default:
						fmt.Println("def")
					}

				}
			}
		}()
	})
	btnCreate := widget.NewButton("Create Room", func() {
		go func() {
			if createGameRoom() {
				label1 := widget.NewLabel(session.RoomID)
				label2 := widget.NewLabel(username)
				label3 := widget.NewLabel("")
				btnCopy := widget.NewButton("Copy roomID", func() {
					fyne.Clipboard.SetContent(window.Clipboard(), session.RoomID)
				})
				leaveBtn := widget.NewButton("Quit session", func() {
					go func() {
						if leaveSession() {
							window.SetContent(contentMain)
							return
						}
					}()
				})
				iconRes, err := fyne.LoadResourceFromPath("./assets/circle.png")
				if err != nil {
					fmt.Println(err)
				}
				content1 := container.NewVBox(container.NewCenter(container.NewHBox(label1, widget.NewIcon(iconRes), label2, widget.NewIcon(iconRes), label3)), container.NewCenter(container.NewHBox(leaveBtn, btnCopy)))
				content2 := container.NewAdaptiveGrid(3, widget.NewButton("X", nil), widget.NewButton("X", nil), widget.NewButton("X", nil), widget.NewButton("X", nil), widget.NewButton("X", nil), widget.NewButton("X", nil), widget.NewButton("X", nil), widget.NewButton("X", nil), widget.NewButton("X", nil))
				window.SetContent(container.NewBorder(content1, nil, nil, nil, content2))

				messageChanHost := make(chan Message)
				for { //
					go receiveMess(messageChanHost)

					t := <-messageChanHost
					fmt.Println(t)
					switch t.Request {
					case "gameJoin":
						fmt.Println("game join")
						label3.SetText(t.Content)
						content1.Refresh()
					case "sessionLeave":
						fmt.Println("session Leave")
						label3.SetText("")
						content1.Refresh()
						return
					case "sessionDisbanded":
						fmt.Println("sessionDisband")
						return
					default:
						fmt.Println("def")
					}
				}
			}
		}()
	})
	contentMain = container.NewCenter(container.NewVBox(roomID, btnJoin, btnCreate))
	window.SetContent(contentMain)
	window.Show()

}

func leaveSession() bool { //wiado z successem musi byc w game status updates
	fmt.Println("num of goroutines: ", runtime.NumGoroutine())

	message := Message{userID, "leave", session.RoomID}
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

	err = json.Unmarshal(data, &message)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if message.Content == "fail" {
		return false
	}

	return true
}

func gameStatusUpdates(c chan Message, container *fyne.Container) {
	//listen for game changes
	//user leaves
	//game progress

	//recive message from server message := Message{server, "leave", username}  //userID -> user who left  //this will be in gameStatusUpdates

	// t := <-c
	// switch t.Request {
	// case "gameJoin":
	// 	fmt.Println("game join")
	// 	container.Add(widget.NewLabel(t.Content))
	// 	container.Refresh()
	// case "sessionLeave":
	// 	fmt.Println("session Leave")

	// default:
	// 	fmt.Println("def")
	// }
}

func receiveMess(c chan Message) {
	fmt.Println("num of goroutines: ", runtime.NumGoroutine())

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

	fmt.Println("num of goroutines: ", runtime.NumGoroutine())
}
