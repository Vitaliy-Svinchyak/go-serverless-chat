package lib

import (
	"fmt"
	"log"
	"strings"
	"github.com/jroimartin/gocui"
	"encoding/json"
	"io/ioutil"
)

var userName = ""
var gu *gocui.Gui
var userList []string
var messages []Message

type Message struct {
	Who     string
	Message string
}

type Label struct {
	name string
	x, y int
	w, h int
	body string
}

type UserList struct {
	Users []string
	x, y  int
	w, h  int
}

func PaintGui(user string) {
	userName = user
	g, err := gocui.NewGui(gocui.OutputNormal)
	gu = g
	maxX, maxY := g.Size()
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true

	label := newLabel("label", 0, 0, "")
	input := newInput("input", 0, maxY-3, maxX-3, 50)
	userList := newUserList(maxX-25, 0)
	focus := gocui.ManagerFunc(setFocus("input"))
	g.SetManager(label, userList, input, focus)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, handleEnter); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func newUserList(x, y int) *UserList {
	w := 22
	h := 10

	return &UserList{x: x, y: y, w: w, h: h}
}

func (l *UserList) Layout(g *gocui.Gui) error {
	v, err := g.SetView("userList", l.x, l.y, l.x+l.w, l.y+l.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		fmt.Fprint(v, "Users online:\n")
		fmt.Fprint(v, strings.Join(userList, ", "))
	} else {
		return err
	}

	return nil
}

func newLabel(name string, x, y int, body string) *Label {
	lines := strings.Split(body, "\n")

	w := 250
	h := 10
	for _, l := range lines {
		if len(l) > w {
			w = len(l)
		}
	}
	//h := len(lines) + 1
	w = w + 1

	return &Label{name: name, x: x, y: y, w: w, h: h, body: body}
}

func (l *Label) Layout(g *gocui.Gui) error {
	v, err := g.SetView(l.name, l.x, l.y, l.x+l.w, l.y+l.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		fmt.Fprint(v, l.body)
	}
	return nil
}

type Input struct {
	name      string
	x, y      int
	w         int
	maxLength int
}

func newInput(name string, x, y, w, maxLength int) *Input {
	return &Input{name: name, x: x, y: y, w: w, maxLength: maxLength}
}

func (i *Input) Layout(g *gocui.Gui) error {
	v, err := g.SetView(i.name, i.x, i.y, i.x+i.w, i.y+2)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editor = i
		v.Editable = true
	}
	return nil
}

var inputBuff []rune

func (i *Input) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	cx, _ := v.Cursor()
	ox, _ := v.Origin()
	limit := ox+cx+1 > i.maxLength

	switch {
	case ch != 0 && mod == 0 && !limit:
		inputBuff = append(inputBuff, ch)
		v.EditWrite(ch)
	case key == gocui.KeySpace && !limit:
		v.EditWrite(' ')
		inputBuff = append(inputBuff, ' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	}
}

func setFocus(name string) func(g *gocui.Gui) error {
	return func(g *gocui.Gui) error {
		_, err := g.SetCurrentView(name)
		return err
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	SetUserOffline()
	return gocui.ErrQuit
}

func handleEnter(g *gocui.Gui, view *gocui.View) error {
	var message = Message{Who: userName, Message: string(inputBuff)}
	messages = append(messages, message)
	SendMessage(message)
	renderMessages()

	return nil
}

func getMessagesBuff() string {
	var inputBuffTemp []string

	for _, message := range messages {
		inputBuffTemp = append(inputBuffTemp, strings.Join([]string{message.Who, message.Message}, ": "))
	}

	return strings.Join(inputBuffTemp, "\n")
}

func NewMessage(message []byte) {
	var messObj = Message{}
	json.Unmarshal(message, &messObj)
	messages = append(messages, messObj)
	gu.Update(func(g *gocui.Gui) error {
		renderMessages()

		return nil
	})
}

func renderMessages() {
	v, _ := gu.View("label")
	v.Clear()
	fmt.Fprintln(v, getMessagesBuff())
	vIn, _ := gu.View("input")
	vIn.Clear()
	vIn.SetCursor(0, 0)
	inputBuff = []rune{}
}

func renderUsers() {
	v, _ := gu.View("userList")
	v.Clear()
	v.Rewind()
}

func writeTofile(d1 []byte) {
	ioutil.WriteFile("./debug", d1, 0644)
}

func SetUserList(users []string) {
	userList = users

	gu.Update(func(g *gocui.Gui) error {
		renderUsers()

		return nil
	})
}
