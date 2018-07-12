package main

import (
	"os"
	"chatting/lib"
)

var messages []string

func main() {
	argsWithProg := os.Args
	var userName = argsWithProg[1]
	lib.ConnectUser(userName)

	lib.PaintGui(userName)
}
