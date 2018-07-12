package lib

import (
	"net"
	"log"
	"fmt"
	"strings"
	"net/http"
	"bytes"
	"io/ioutil"
	"encoding/json"
	"strconv"
	"time"
)

type userIp struct {
	Name   string
	Ip     string
	Online bool
	Id     int
	IdDb   string `json:"_id"`
}

var users []userIp
var currentUser userIp

func ConnectUser(userName string) {
	var users = GetUsers()
	var ip = GetOutboundIP()
	connected := false

	for _, user := range users {
		if user.Ip == ip {
			currentUser = user
			connected = true
			setUserOnline(user)
			break
		}
	}

	if !connected {
		insertUser(userName)
	}

	ConnectUserToSocket()
	actualizeUsersList()
}

func GetCachedUsers() []userIp {
	return users
}

func GetUsers() []userIp {
	fmt.Println(`Getting a list of users`)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://testit-abb9.restdb.io/rest/userips", nil)
	req.Header.Add("cache-control", `no-cache`)
	req.Header.Add("x-apikey", `dd5ea8826752ead9ee67a5ad21be7e43b501f`)
	req.Header.Add("content-type", `application/json`)
	r, _ := client.Do(req)
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &users)

	return users
}

func insertUser(userName string) {
	fmt.Println(`Creating new user`)
	var ip = GetOutboundIP()
	s := []string{`{"name":"`, userName, `","ip":"`, ip, `","online":true}`}
	var user = []byte(strings.Join(s, ""))

	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://testit-abb9.restdb.io/rest/userips", bytes.NewBuffer(user))
	req.Header.Add("cache-control", `no-cache`)
	req.Header.Add("x-apikey", `dd5ea8826752ead9ee67a5ad21be7e43b501f`)
	req.Header.Add("content-type", `application/json`)
	client.Do(req)
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func setUserOnline(user userIp) {
	fmt.Println(`Setting online`)
	setUserOnlineStatus(user, true)
}

func SetUserOffline() {
	SendMessage(Message{Who: currentUser.Name, Message: "left chat"})
	setUserOnlineStatus(currentUser, false)
}

func setUserOnlineStatus(user userIp, status bool) {
	var jsonStr = []byte(strings.Join([]string{`{"online":`, strconv.FormatBool(status), `}`}, ""))
	var url = strings.Join([]string{"https://testit-abb9.restdb.io/rest/userips/", user.IdDb}, "")

	client := &http.Client{}
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))
	req.Header.Add("cache-control", `no-cache`)
	req.Header.Add("x-apikey", `dd5ea8826752ead9ee67a5ad21be7e43b501f`)
	req.Header.Add("content-type", `application/json`)
	r, _ := client.Do(req)
	ioutil.ReadAll(r.Body)
}

func actualizeUsersList() {
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				var users = GetUsers()
				var userList []string
				for _, user := range users {
					userList = append(userList, user.Name)
				}
				SetUserList(userList)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
