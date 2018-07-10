package db

import (
	"net"
	"log"
	"fmt"
	"strings"
	"net/http"
	"bytes"
	"io/ioutil"
	"encoding/json"
)

type userIp struct {
	Name   string
	Ip     string
	Online bool
	Id     int
	IdDb   string `json:"_id"`
}

func ConnectUser(userName string) {
	var users = getUsers()
	var ip = GetOutboundIP()
	connected := false

	for _, user := range users {
		if user.Ip == ip {
			connected = true
			setUserOnline(user)
			break
		}
	}

	if !connected {
		insertUser(userName)
	}
}

func getUsers() []userIp {
	fmt.Println(`Getting a list of users`)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://testit-abb9.restdb.io/rest/userips", nil)
	req.Header.Add("cache-control", `no-cache`)
	req.Header.Add("x-apikey", `dd5ea8826752ead9ee67a5ad21be7e43b501f`)
	req.Header.Add("content-type", `application/json`)
	r, _ := client.Do(req)
	body, _ := ioutil.ReadAll(r.Body)
	var data []userIp
	json.Unmarshal(body, &data)

	return data
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
	//body, _ := ioutil.ReadAll(r.Body)
	//fmt.Println(body)
}

// Get preferred outbound ip of this machine
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
	fmt.Println(`Seting online`)
	//s := []string{`{"name":"`, user.Name, `","ip":"`, user.Ip, `","online":true}`}
	//var jsonStr = []byte(strings.Join(s, ""))
	var jsonStr = []byte(`{"online":true}`)
	var url = strings.Join([]string{"https://testit-abb9.restdb.io/rest/userips/", user.IdDb}, "")

	client := &http.Client{}
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))
	req.Header.Add("cache-control", `no-cache`)
	req.Header.Add("x-apikey", `dd5ea8826752ead9ee67a5ad21be7e43b501f`)
	req.Header.Add("content-type", `application/json`)
	r, _ := client.Do(req)
	ioutil.ReadAll(r.Body)
}
