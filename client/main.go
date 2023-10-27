package main

import (
	"encoding/json"
	"fmt"
	"log"
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
var content *fyne.Container
var btns [][]*widget.Button
var iconTurnClient *widget.Icon
var iconTurnHost *widget.Icon
var playerShape string
var popUp *widget.PopUp

type Session struct {
	RoomID  string
	Players []string
	Host    string
	Turn    string
}

type Message struct {
	Sender  string
	Request string
	Content map[string]interface{}
	Session Session
}

func (m Message) send() {
	data, err := json.Marshal(m)
	if err != nil {
		fmt.Println("func send", err)
		return
	}

	_, err = connection.Write(data)
	if err != nil {
		fmt.Println("func send", err)
		return
	}
}

func connectToServer() bool {
	message := Message{userID, "register", map[string]interface{}{"username": username}, session}

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
			btns[i][j].SetText(playerShape)

			ij := strconv.Itoa(i) + "," + strconv.Itoa(j)
			message := Message{userID, "status", map[string]interface{}{"roomID": session.RoomID, "move": ij, "player": playerShape}, session}
			message.send()

			fmt.Println(session.Players)
		}
	}
}

func initializeGameWindow() {
	window := myApp.NewWindow("Tic-Tac-Toe") //
	window.Resize(fyne.NewSize(600, 600))
	roomID := widget.NewEntry()

	btnJoin := widget.NewButton("Join Room", func() {
		go func() {
			fmt.Println("num of goroutines: func: btnJoin", runtime.NumGoroutine())
			if joinRoom(roomID.Text, window) {
				playerShape = "O"
				messageChan := make(chan Message)
				label1 := widget.NewLabel(session.RoomID)
				label2 := widget.NewLabel(session.Players[0])
				label3 := widget.NewLabel(session.Players[1])
				leaveBtn := widget.NewButton("Quit session", func() {
					go func() {
						if leaveSession() {
							window.SetContent(contentMain)
							fmt.Println("num of goroutines: func: in leavesession if", runtime.NumGoroutine())
							runtime.Goexit()
						}
					}()
					fmt.Println("num of goroutines: func: after leavesession", runtime.NumGoroutine())
				})
				content = container.NewVBox(container.NewCenter(container.NewHBox(label1, iconTurnHost, label2, label3, iconTurnClient)), container.NewCenter(container.NewHBox(leaveBtn)))
				btns = [][]*widget.Button{
					{widget.NewButton("", func() { gameBtnTapped(0, 0) }), widget.NewButton("", func() { gameBtnTapped(0, 1) }), widget.NewButton("", func() { gameBtnTapped(0, 2) })},
					{widget.NewButton("", func() { gameBtnTapped(1, 0) }), widget.NewButton("", func() { gameBtnTapped(1, 1) }), widget.NewButton("", func() { gameBtnTapped(1, 2) })},
					{widget.NewButton("", func() { gameBtnTapped(2, 0) }), widget.NewButton("", func() { gameBtnTapped(2, 1) }), widget.NewButton("", func() { gameBtnTapped(2, 2) })},
				}

				contentFinal := container.NewBorder(content, nil, nil, nil)

				content = container.NewAdaptiveGrid(3)
				for i := 0; i < 3; i++ {
					for j := 0; j < 3; j++ {
						content.Add(btns[i][j])
					}
				}

				contentFinal.Add(content)
				window.SetContent(contentFinal)

				for {
					go receiveMess(messageChan)

					gameStatusUpdates(messageChan, window)

					if userID == session.Turn {
						iconTurnClient.Show()
						iconTurnHost.Hide()
					} else {
						iconTurnClient.Hide()
						iconTurnHost.Show()
					}
				}
			}
		}()
	})
	btnCreate := widget.NewButton("Create Room", func() {
		go func() {
			if createGameRoom() {
				playerShape = "X"
				messageChan := make(chan Message)
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
							fmt.Println("num of goroutines: func: in leavesession if", runtime.NumGoroutine())
							runtime.Goexit()
						}
					}()
					fmt.Println("num of goroutines: func: after leavesession", runtime.NumGoroutine())
				})
				content = container.NewVBox(container.NewCenter(container.NewHBox(label1, iconTurnHost, label2, label3, iconTurnClient)), container.NewCenter(container.NewHBox(leaveBtn, btnCopy)))
				btns = [][]*widget.Button{
					{widget.NewButton("", func() { gameBtnTapped(0, 0) }), widget.NewButton("", func() { gameBtnTapped(0, 1) }), widget.NewButton("", func() { gameBtnTapped(0, 2) })},
					{widget.NewButton("", func() { gameBtnTapped(1, 0) }), widget.NewButton("", func() { gameBtnTapped(1, 1) }), widget.NewButton("", func() { gameBtnTapped(1, 2) })},
					{widget.NewButton("", func() { gameBtnTapped(2, 0) }), widget.NewButton("", func() { gameBtnTapped(2, 1) }), widget.NewButton("", func() { gameBtnTapped(2, 2) })},
				}

				contentFinal := container.NewBorder(content, nil, nil, nil)

				content = container.NewAdaptiveGrid(3)
				for i := 0; i < 3; i++ {
					for j := 0; j < 3; j++ {
						content.Add(btns[i][j])
					}
				}
				contentFinal.Add(content)

				window.SetContent(contentFinal)

				for {
					go receiveMess(messageChan)

					gameStatusUpdates(messageChan, window)

					if userID == session.Turn {
						iconTurnClient.Hide()
						iconTurnHost.Show()
					} else {
						iconTurnClient.Show()
						iconTurnHost.Hide()
					}
				}
			}
		}()
	})
	btnExit := widget.NewButton("Exit", func() {
		message := Message{userID, "disconnect", map[string]interface{}{}, session}
		message.send()

		connection.Close()
		myApp.Quit()
	})
	contentMain = container.NewCenter(container.NewVBox(roomID, btnJoin, btnCreate, btnExit))
	window.SetContent(contentMain)
	window.Show()

}

func leaveSession() bool {
	fmt.Println("num of goroutines: func: leavesession", runtime.NumGoroutine())

	message := Message{userID, "leaveSession", map[string]interface{}{"roomID": session.RoomID}, session}
	message.send()

	return true
}

func gameStatusUpdates(c chan Message, w fyne.Window) {
	t := <-c
	fmt.Println("func gamestatusupdates", t)
	switch t.Request {
	case "gameJoin":
		fmt.Println("game join")
		c := w.Content().(*fyne.Container)
		c.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[3].(*widget.Label).SetText(t.Content["username"].(string))
		fmt.Println(session)
		session = t.Session
		fmt.Println(session)
		w.Content().Refresh()
	case "sessionLeave":
		fmt.Println("session Leave")
		c := w.Content().(*fyne.Container)
		c.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[3].(*widget.Label).SetText("")
		fmt.Println(session)
		session = t.Session
		fmt.Println(session)
		popUp.Hide()
		w.Content().Refresh()
	case "sessionDisbanded":
		fmt.Println("sessionDisband")
		w.SetContent(contentMain)
		session = Session{}
		popUp.Hide()
		close(c)
		runtime.Goexit()
	case "leave":
		fmt.Println("leave ", t.Content)
		if t.Content["status"].(string) == "success" {
			fmt.Println("success")
			w.SetContent(contentMain)
			session = Session{}
		}
		close(c)
		runtime.Goexit()
	case "status":
		fmt.Println("status", t.Content)
		fmt.Println(session)
		session = t.Session
		fmt.Println(session)

		_, gameEnd := t.Content["gameEnd"]
		if gameEnd {
			_, draw := t.Content["draw"]
			if draw {
				displayEndGamePopUp("draw", w)
			} else {
				displayEndGamePopUp("winner: "+t.Content["winner"].(string), w)
			}
		}

		_, move := t.Content["move"]
		if move {
			ij := strings.Split(t.Content["move"].(string), ",")
			i, _ := strconv.Atoi(ij[0])
			j, _ := strconv.Atoi(ij[1])
			btns[i][j].SetText(t.Content["player"].(string))

			if checkHorizontal() || checkVertical() || checkDiagonal() {
				message := Message{userID, "status", map[string]interface{}{"roomID": session.RoomID, "gameEnd": true, "winner": playerShape}, session}
				message.send()
			}

			if checkEmpty() {
				message := Message{userID, "status", map[string]interface{}{"roomID": session.RoomID, "gameEnd": true, "draw": ""}, session}
				message.send()
			}
		}
	case "rematch":
		fmt.Println("rematch")
		popUp.Hide()
		var rematchPopUp *widget.PopUp
		popUpContent := container.NewVBox(widget.NewLabel("Rematch?"), widget.NewButton("YES", func() {
			message := Message{userID, "newGame", map[string]interface{}{}, session}
			message.send()
			rematchPopUp.Hide()
		}),
			widget.NewButton("EXIT", func() {
				rematchPopUp.Hide()
				go func() {
					if leaveSession() {
						w.SetContent(contentMain)
						fmt.Println("num of goroutines: func: in leavesession if", runtime.NumGoroutine())
						runtime.Goexit()
					}
				}()
			}))
		rematchPopUp = widget.NewModalPopUp(popUpContent, w.Canvas())
		rematchPopUp.Show()
	case "newGame":
		log.Println("newGame")
		if !popUp.Hidden {
			popUp.Hide()
		}
		btns := w.Content().(*fyne.Container).Objects[1].(*fyne.Container).Objects
		for _, btn := range btns {
			btn.(*widget.Button).SetText("")
		}
		log.Println(btns)
	default:
		fmt.Println("def gamestatusupdates")
	}
}

func displayEndGamePopUp(text string, w fyne.Window) {
	content = container.NewVBox(widget.NewLabel(text),
		widget.NewButton("REMATCH", func() {
			message := Message{userID, "status", map[string]interface{}{"roomID": session.RoomID, "gameEnd": true, "rematch": true}, session} //
			message.send()
			content.Objects[1].(*widget.Button).SetText("WAITING FOR OTHER PLAYER")
		}),
		widget.NewButton("EXIT SESSION", func() {
			popUp.Hide()
			go func() {
				if leaveSession() {
					w.SetContent(contentMain)
					fmt.Println("num of goroutines: func: in leavesession if", runtime.NumGoroutine())
					runtime.Goexit()
				}
			}()
		}))
	popUp = widget.NewModalPopUp(content, w.Canvas())
	popUp.Show()
}

func checkEmpty() bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if btns[j][i].Text == "" {
				fmt.Println("emppty")
				return false
			}
		}
	}
	fmt.Println("no emppty left")
	return true
}

func checkDiagonal() bool {
	if (btns[0][0].Text == playerShape && btns[1][1].Text == playerShape && btns[2][2].Text == playerShape) || (btns[0][2].Text == playerShape && btns[1][1].Text == playerShape && btns[2][0].Text == playerShape) {
		fmt.Println("won")
		return true
	}
	return false
}

func checkVertical() bool {
	c := 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if btns[j][i].Text == playerShape {
				c++
				fmt.Println(c)
				if c == 3 {
					fmt.Println(playerShape, "won")
					return true
				}
			}
		}
		c = 0
	}
	return false
}

func checkHorizontal() bool {
	c := 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if btns[i][j].Text == playerShape {
				c++
				fmt.Println(c)
				if c == 3 {
					fmt.Println(playerShape, "won")
					return true
				}
			}
		}
		c = 0
	}
	return false
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

func joinRoom(roomID string, w fyne.Window) bool {
	message := Message{userID, "joinRoom", map[string]interface{}{"roomID": roomID}, session}
	message.send()

	data := make([]byte, 1024)
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

	if session.RoomID == "" {
		fmt.Println("COULDN'T JOIN THE SESSION")
		content = container.NewVBox(widget.NewLabel("COULDN'T JOIN THE SESSION"), widget.NewButton("OK", func() { //
			popUp.Hide()
		}))
		popUp = widget.NewModalPopUp(content, w.Canvas())
		popUp.Show()
		return false
	}

	fmt.Println("JOIn success")
	return true
}

func createGameRoom() bool {
	fmt.Println("func creategameroom", connection)

	message := Message{userID, "createRoom", map[string]interface{}{}, session}
	message.send()

	data := make([]byte, 1024)
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

	iconRes, err := fyne.LoadResourceFromPath("./assets/circle.png")
	if err != nil {
		fmt.Println("func bntJoin", err)
	}
	iconTurnClient = widget.NewIcon(iconRes)
	iconTurnClient.Hide()
	iconTurnHost = widget.NewIcon(iconRes)

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
