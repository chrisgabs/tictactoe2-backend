package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Player struct {
	Conn *websocket.Conn
	id   float64
}

var connectedPlayers = []Player{}

// func (p Player) broadCastMousePosition(data []byte) {
// 	// toSend, err := json.Marshal(data)
// 	// if err != nil {
// 	// 	log.Println("error in marshaling broadcast message data")
// 	// }

// 	if err := p.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
// 		log.Println(err)
// 		return
// 	}
// }

type messageData struct {
	ClientX      float64
	ClientY      float64
	PlayerNumber float64
	PieceID      string
}

var numConnections float64 = 0.0

func sampleEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "sample endpoint working")
}

func sendToAllExceptSelf(data []byte) {
	var msgData messageData
	err := json.Unmarshal(data, &msgData)
	if err != nil {
		fmt.Println("error in unmrashalling")
		log.Println(err)
		return
	}
	for _, player := range connectedPlayers {
		if player.id != msgData.PlayerNumber {
			if err := player.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}

	defer ws.Close()

	p := Player{
		ws,
		numConnections,
	}

	defer func() {
		fmt.Printf("disconnected connection id: %v\n", p.id)
	}()

	connectedPlayers = append(connectedPlayers, p)
	numConnections += 1

	log.Println("Client Connected")

	// infinite loop
	for {
		messageType, data, err := p.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		if messageType == websocket.TextMessage {
			fmt.Printf("From client: %v | %v\n", string(data), p.id)
		}

		sendToAllExceptSelf(data)
	}
}

func setupRoutes() {
	http.HandleFunc("/", sampleEndpoint)
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	fmt.Println("hello world")
	setupRoutes()

	http.ListenAndServe(":8080", nil)
}
