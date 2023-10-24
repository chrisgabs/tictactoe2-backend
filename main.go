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
	Id   float64
}

// TODO: Implement this as a map instead of slice
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

type MoveData struct {
	ClientX      float64
	ClientY      float64
	PlayerNumber float64
	PieceID      string
}

type MessageData struct {
	EventType string
	Data interface{}
}

var numConnections float64 = 0.0

func sampleEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "sample endpoint working")
}

func sendToAllExceptSelf(data MessageData) {
	move := data.Data.(MoveData)
	for _, player := range connectedPlayers {
		if player.Id != move.PlayerNumber {
			if err := player.Conn.WriteJSON(data); err != nil {
				log.Println("error sending json 62")
				log.Println(err)
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

	defer func () {
		var updatedSlice = []Player{}
		for _, player := range connectedPlayers {
			if player.Id != p.Id {
				updatedSlice = append(updatedSlice, player)
			}
		}
		connectedPlayers = updatedSlice;
	}()
	
	defer func() {
		fmt.Printf("disconnected id: %v\n", p.Id)
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

		
		// parse message
		var message MessageData
		if err1 := json.Unmarshal(data, &message); err1 != nil {
			log.Println("error in unmrashalling line 106")
			log.Println(err1)
			return
		}
		
		fmt.Printf("| %v | %v | %v \n", message.EventType, messageType, message)

		// check what to do based on message type
		switch eType := message.EventType; eType {
			case "connect":
				response := MessageData {"connect", p,}
				if err := ws.WriteJSON(response); err != nil {
					fmt.Println("error in writing JSON 120")
				}
			case "move":
				moveData, _ := message.Data.(map[string]interface{})
				fmt.Println(moveData)
				move := MoveData{
					ClientX:      moveData["ClientX"].(float64),
					ClientY:      moveData["ClientY"].(float64),
					PlayerNumber: moveData["PlayerNumber"].(float64),
					PieceID:      moveData["PieceID"].(string),
				}
				response := MessageData{"move", move}
				sendToAllExceptSelf(response)
		}
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
