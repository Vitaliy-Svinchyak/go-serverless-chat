package lib

import (
	"fmt"
	"net"
	"strings"
	"bytes"
	"encoding/json"
)

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	socket net.Conn
	data   chan []byte
}

var connections = map[string]net.Conn{}

func (manager *ClientManager) start() {
	for {
		select {
		case connection := <-manager.register:
			manager.clients[connection] = true
		case connection := <-manager.unregister:
			if _, ok := manager.clients[connection]; ok {
				close(connection.data)
				delete(manager.clients, connection)
			}
		case message := <-manager.broadcast:
			for connection := range manager.clients {
				select {
				case connection.data <- message:
				default:
					close(connection.data)
					delete(manager.clients, connection)
				}
			}
		}
	}
}

func (manager *ClientManager) receive(client *Client) {
	for {
		message := make([]byte, 4096)
		length, err := client.socket.Read(message)
		if err != nil {
			manager.unregister <- client
			client.socket.Close()
			break
		}
		message = bytes.Trim(message, "\x00")
		if length > 0 {
			NewMessage(message)
			//fmt.Println("RECEIVED: " + string(message))
			manager.broadcast <- message
		}
	}
}

func (client *Client) receive() {
	for {
		message := make([]byte, 4096)
		length, err := client.socket.Read(message)
		if err != nil {
			client.socket.Close()
			break
		}
		message = bytes.Trim(message, "\x00")
		if length > 0 {
			//fmt.Println("RECEIVED: " + string(message))
		}
	}
}

func (manager *ClientManager) send(client *Client) {
	defer client.socket.Close()
	for {
		select {
		case message, ok := <-client.data:
			if !ok {
				return
			}
			client.socket.Write(message)
		}
	}
}

func StartServer() {
	listener, error := net.Listen("tcp", ":12345")
	if error != nil {
		fmt.Println(error)
	}
	manager := ClientManager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go manager.start()

	go func() {
		for {
			connection, _ := listener.Accept()
			if error != nil {
				fmt.Println(error)
			}

			client := &Client{socket: connection, data: make(chan []byte)}
			manager.register <- client
			go manager.receive(client)
			go manager.send(client)
		}
	}()
}

func ConnectToTheServer(host string) {
	connection, error := net.Dial("tcp", strings.Join([]string{host, "12345"}, ":"))
	connections[host] = connection
	if error != nil {
		fmt.Println(error)
	}
	client := &Client{socket: connection}
	go client.receive()
}

func ConnectUserToSocket() {
	StartServer()
	var users = GetCachedUsers()

	for _, user := range users {
		if user.Online {
			fmt.Println("Connecting", user)
			ConnectToTheServer(user.Ip)
		}
	}
}

func SendMessage(message Message) {
	for _, connection := range connections {
		messageJson, _ := json.Marshal(message)
		connection.Write([]byte(strings.TrimRight(string(messageJson), "\n")))
	}
}
