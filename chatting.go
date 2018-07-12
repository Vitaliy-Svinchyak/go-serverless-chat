package main

import (
	"os"
	"chatting/lib"
)

var messages []string

func main() {
	argsWithProg := os.Args
	var userName = argsWithProg[1]

	//socket.ConnectToTheServer(userName)
	//socket.StartServer(userName)
	lib.ConnectUser(userName)

	lib.PaintGui(userName)
}
