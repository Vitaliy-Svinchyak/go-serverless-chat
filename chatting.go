package main

import (
	//"chatting/db"
	"chatting/socket"
	"os"
	"chatting/gui"
)

var messages []string

func main() {
	argsWithProg := os.Args
	var userName = argsWithProg[1]

	socket.ConnectToTheServer(userName)
	//socket.StartServer(userName)

	//db.ConnectUser(userName)
	gui.PaintGui(userName)
}
