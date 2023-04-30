package main

import (
	"encoding/json"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"

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
var contentMain *fyne.Container
var content1 *fyne.Container
var content2 *fyne.Container
var btns [][]*widget.Button

type Session struct {
	RoomID  string
	Players []string //
	Host    string
	Turn    string
}

type Message struct {
	Sender  string
	Request string
	Content string
	Session Session
}

func connectToServer() bool {
	message := Message{userID, "register", username, session}

	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println("func connect", err)
		return false
	}

	connection, err = net.Dial("tcp", ":8080")
	if err != nil {
		fmt.Println("func connect", err)
		return false
	}
	// defer connection.Close() //need opened connection for later actions

	_, err = connection.Write(data)
	if err != nil {
		fmt.Println("func connect", err)
		return false
	}

	data = make([]byte, 6)
	_, err = connection.Read(data)
	if err != nil {
		fmt.Println("func connect", err)
	}

	userID = string(data)

	fmt.Println("func connect", userID)
	fmt.Println("func connect", username)
	return true
}

func gameBtnTapped(i, j int) {
	if userID == session.Turn {
		if btns[i][j].Text == "" {
			btns[i][j].SetText("X")

			ij := strconv.Itoa(i) + "," + strconv.Itoa(j)
			message := Message{userID, "statuschange", session.RoomID + "|" + ij, session}

			data, err := json.Marshal(message)
			if err != nil {
				fmt.Println("func gameBtnTapped", err)
			}

			_, err = connection.Write(data)
			if err != nil {
				fmt.Println("func gameBtnTapped", err)
			}

			fmt.Println(session.Players)
		}
	}
}

func initializeGameWindow() {
	window := myApp.NewWindow("Tic-Tac-Toe")
	window.Resize(fyne.NewSize(600, 600))
	roomID := widget.NewEntry()

	btnJoin := widget.NewButton("Join Room", func() {
		go func() {
			fmt.Println("num of goroutines: func: btnJoin", runtime.NumGoroutine())
			if joinRoom(roomID.Text) {
				messageChan := make(chan Message)
				label1 := widget.NewLabel(session.RoomID)
				label2 := widget.NewLabel(session.Players[0])
				label3 := widget.NewLabel(session.Players[1])
				leaveBtn := widget.NewButton("Quit session", func() { //
					go func() {
						if leaveSession() {
							window.SetContent(contentMain)
							fmt.Println("num of goroutines: func: in leavesession if", runtime.NumGoroutine())
							runtime.Goexit()
						}
					}()
					fmt.Println("num of goroutines: func: after leavesession", runtime.NumGoroutine())
				})
				iconRes, err := fyne.LoadResourceFromPath("./assets/circle.png")
				if err != nil {
					fmt.Println("func bntJoin", err)
				}
				content1 = container.NewVBox(container.NewCenter(container.NewHBox(label1, widget.NewIcon(iconRes), label2, widget.NewIcon(iconRes), label3)), container.NewCenter(container.NewHBox(leaveBtn)))
				btns = [][]*widget.Button{
					{widget.NewButton("", func() { gameBtnTapped(0, 0) }), widget.NewButton("", func() { gameBtnTapped(0, 1) }), widget.NewButton("", func() { gameBtnTapped(0, 2) })},
					{widget.NewButton("", func() { gameBtnTapped(1, 0) }), widget.NewButton("", func() { gameBtnTapped(1, 1) }), widget.NewButton("", func() { gameBtnTapped(1, 2) })},
					{widget.NewButton("", func() { gameBtnTapped(2, 0) }), widget.NewButton("", func() { gameBtnTapped(2, 1) }), widget.NewButton("", func() { gameBtnTapped(2, 2) })},
				}

				content2 = container.NewAdaptiveGrid(3)
				for i := 0; i < 3; i++ {
					for j := 0; j < 3; j++ {
						content2.Add(btns[i][j])
					}
				}
				window.SetContent(container.NewBorder(content1, nil, nil, nil, content2))

				for {
					go receiveMess(messageChan)

					gameStatusUpdates(messageChan, label3, window)
				}
			}

		}()
	})
	btnCreate := widget.NewButton("Create Room", func() {
		go func() {
			if createGameRoom() {
				messageChan := make(chan Message)
				label1 := widget.NewLabel(session.RoomID)
				label2 := widget.NewLabel(username)
				label3 := widget.NewLabel("")
				btnCopy := widget.NewButton("Copy roomID", func() {
					fyne.Clipboard.SetContent(window.Clipboard(), session.RoomID)
				})
				leaveBtn := widget.NewButton("Quit session", func() { //
					go func() {
						if leaveSession() {
							window.SetContent(contentMain)
							fmt.Println("num of goroutines: func: in leavesession if", runtime.NumGoroutine())
							runtime.Goexit()
						}
					}()
					fmt.Println("num of goroutines: func: after leavesession", runtime.NumGoroutine())
				})
				iconRes, err := fyne.LoadResourceFromPath("./assets/circle.png")
				if err != nil {
					fmt.Println("func btnCreate", err)
				}
				content1 = container.NewVBox(container.NewCenter(container.NewHBox(label1, widget.NewIcon(iconRes), label2, widget.NewIcon(iconRes), label3)), container.NewCenter(container.NewHBox(leaveBtn, btnCopy)))
				btns = [][]*widget.Button{
					{widget.NewButton("", func() { gameBtnTapped(0, 0) }), widget.NewButton("", func() { gameBtnTapped(0, 1) }), widget.NewButton("", func() { gameBtnTapped(0, 2) })},
					{widget.NewButton("", func() { gameBtnTapped(1, 0) }), widget.NewButton("", func() { gameBtnTapped(1, 1) }), widget.NewButton("", func() { gameBtnTapped(1, 2) })},
					{widget.NewButton("", func() { gameBtnTapped(2, 0) }), widget.NewButton("", func() { gameBtnTapped(2, 1) }), widget.NewButton("", func() { gameBtnTapped(2, 2) })},
				}

				content2 = container.NewAdaptiveGrid(3)
				for i := 0; i < 3; i++ {
					for j := 0; j < 3; j++ {
						content2.Add(btns[i][j])
					}
				}
				window.SetContent(container.NewBorder(content1, nil, nil, nil, content2))

				for {
					go receiveMess(messageChan)

					gameStatusUpdates(messageChan, label3, window)
				}
			}
		}()
	})
	btnExit := widget.NewButton("Exit", func() {
		message := Message{userID, "dc", "", session}
		data, err := json.Marshal(message)
		if err != nil {
			fmt.Println("func leaveSession", err)
			return
		}

		_, err = connection.Write(data)
		if err != nil {
			fmt.Println("func leavesession", err)
			return
		}
		connection.Close()
		myApp.Quit()
	})
	contentMain = container.NewCenter(container.NewVBox(roomID, btnJoin, btnCreate, btnExit))
	window.SetContent(contentMain)
	window.Show()

}

func leaveSession() bool {
	fmt.Println("num of goroutines: func: leavesession", runtime.NumGoroutine())

	message := Message{userID, "leave", session.RoomID, session}
	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println("func leaveSession", err)
		return false
	}

	_, err = connection.Write(data)
	if err != nil {
		fmt.Println("func leavesession", err)
		return false
	}

	return true
}

func gameStatusUpdates(c chan Message, l *widget.Label, w fyne.Window) {
	t := <-c
	fmt.Println("func gamestatusupdates", t)
	switch t.Request {
	case "gameJoin":
		fmt.Println("game join")
		l.SetText(t.Content)
		fmt.Println(session)
		session = t.Session
		fmt.Println(session)
		content1.Refresh()
	case "sessionLeave":
		fmt.Println("session Leave")
		l.SetText("")
		fmt.Println(session)
		session = t.Session
		fmt.Println(session)
		content1.Refresh()
	case "sessionDisbanded":
		fmt.Println("sessionDisband")
		w.SetContent(contentMain)
		session = Session{}
		close(c)
		runtime.Goexit()
	case "leave":
		fmt.Println("leave ", t.Content)
		if t.Content == "success" {
			fmt.Println("success")
			w.SetContent(contentMain)
			session = Session{}
		}
		close(c)
		runtime.Goexit()
	case "statuschange":
		fmt.Println("statuschange", t.Content)
		fmt.Println(session)
		session = t.Session
		fmt.Println(session)
		ij := strings.Split(t.Content, ",")
		i, _ := strconv.Atoi(ij[0])
		j, _ := strconv.Atoi(ij[1])
		btns[i][j].SetText("X")

		//+ checking here if someone won

	default:
		fmt.Println("def gamestatusupdates")
	}
}

func receiveMess(c chan Message) {
	fmt.Println("num of goroutines: func: receiveMess", runtime.NumGoroutine())

	data := make([]byte, 1024)
	n, err := connection.Read(data)
	if err != nil {
		fmt.Println("func receivemess", err)
	}
	data = data[:n]

	var mess Message

	err = json.Unmarshal(data, &mess)
	if err != nil {
		fmt.Println("func receivemess", err)
	}

	fmt.Println("func receivemess", mess)
	c <- mess
}

func joinRoom(roomID string) bool {
	message := Message{userID, "joinRoom", roomID, session}

	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println("func joinroom", err)
		return false
	}
	_, err = connection.Write(data)
	if err != nil {
		fmt.Println("func joinroom", err)
		return false
	}

	data = make([]byte, 1024)
	n, err := connection.Read(data)
	if err != nil {
		fmt.Println("func joinroom", err)
		return false
	}
	data = data[:n]

	err = json.Unmarshal(data, &session)
	if err != nil {
		fmt.Println("func joinroom", err)
		return false
	}

	fmt.Println("JOIn success")
	return true
}

func createGameRoom() bool {
	fmt.Println("func creategameroom", connection)

	message := Message{userID, "createRoom", "", session}

	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println("func creategameroom", err)
		return false
	}
	_, err = connection.Write(data)
	if err != nil {
		fmt.Println("func creategameroom", err)
		return false
	}

	//server sents back roomID, current players

	data = make([]byte, 1024)
	n, err := connection.Read(data)
	if err != nil {
		fmt.Println("func creategameroom", err)
		return false
	}
	data = data[:n]

	err = json.Unmarshal(data, &session)
	if err != nil {
		fmt.Println("func creategameroom", err)
		return false
	}

	fmt.Println("func creategameroom", session)
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
		})),
		widget.NewFormItem("", widget.NewButton("Exit", myApp.Quit)))
	form.Resize(fyne.NewSize(200, 200))
	form.Move(fyne.NewPos(150, 200))

	content := container.NewWithoutLayout(form)
	loginWindow.SetContent(content)

	loginWindow.Show()
	myApp.Run()

	fmt.Println("num of goroutines: func: main", runtime.NumGoroutine())
}
