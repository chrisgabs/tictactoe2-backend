package game

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

type Player struct {
	SessionId      string `json:"sessionId"`
	PlayerNumber   int    `json:"playerNumber"`
	DisplayName    string `json:"displayName"`
	Room           *Room
	ReceiveChannel chan *MessageData
	Conn           *websocket.Conn
}

type MoveData struct {
	ClientX      float64
	ClientY      float64
	PlayerNumber float64
	PieceID      string
}

type MessageData struct {
	EventType    string
	PlayerNumber int
	Data         interface{}
}

type DropData struct {
	PlayerNumber int    `json:"playerNumber"`
	Cell         string `json:"cell"`
	Piece        string `json:"piece"`
	IsValidMove  bool   `json:"isValidMove"`
}

func (p *Player) StartListeningToClient() {

	defer func(Conn *websocket.Conn) {
		err := Conn.Close()
		if err != nil {
			log.Printf("Error closing WS connection of [%v]%v : %v\n", p.SessionId, p.DisplayName, err)
		}
		log.Printf("Closed WS connection of [%v]%v\n", p.SessionId, p.DisplayName)
	}(p.Conn)

	for {
		_, data, err := p.Conn.ReadMessage()
		if err != nil {
			log.Printf("ERROR [%v] Problem in reading message from websocket %v\n", p.SessionId, err)
			return
		}

		// parse message
		var message MessageData
		if err1 := json.Unmarshal(data, &message); err1 != nil {
			log.Printf("error in unmrashalling line 51 %v\n", err1)
		}

		log.Printf("| %v | %v | %v \n", message.EventType, message.PlayerNumber, message)

		// check what to do based on message type
		switch eType := message.EventType; eType {
		case Connect:
			data := make(map[string]interface{})
			data["playerNumber"] = p.PlayerNumber
			data["roomId"] = p.Room.RoomId
			data["boardData"] = p.Room.Board.AsRawMessage()
			response := MessageData{Connect, p.PlayerNumber, data}
			if err := p.Conn.WriteJSON(response); err != nil {
				log.Printf("error in writing JSON 64: %v\n", err)
			}
		case Drop: //
			response := MessageData{Drop, p.PlayerNumber, message.Data}
			jsonString, _ := json.Marshal(message.Data)
			var dropData DropData
			err := json.Unmarshal(jsonString, &dropData)
			if err != nil {
				log.Printf("Error in unmarshalling drop data: %v\n", err)
				return
			}
			p.Room.Receiver <- &response
			p.Room.Board.placePiece(dropData)
		case DragEnd:
			response := MessageData{DragEnd, p.PlayerNumber, message.Data}
			p.Room.Receiver <- &response
		case Move:
			moveData, _ := message.Data.(map[string]interface{})
			move := MoveData{
				ClientX:      moveData["ClientX"].(float64),
				ClientY:      moveData["ClientY"].(float64),
				PlayerNumber: moveData["PlayerNumber"].(float64),
				PieceID:      moveData["PieceID"].(string),
			}
			response := MessageData{Move, p.PlayerNumber, move}
			p.Room.Receiver <- &response
		case Join:
			response := MessageData{Join, p.PlayerNumber, message.Data}
			p.Room.Receiver <- &response
		}
	}

}

// StartListeningToRoom to be called only by the room player is in to ensure that p.ReceiveChannel exists
// TODO: determine when to kill this go routine
//   - create a timer that starts when there is no activity in ReceiveChannel. After N time, send
//     send a health check. if health check fails, then kill the go routine and clean up player.
func (p *Player) StartListeningToRoom() {
	defer func() {
		fmt.Printf("[%v] Stopped listening to room [%v]\n", p.SessionId, p.Room.RoomId)
	}()
	for data := range p.ReceiveChannel {
		if err := p.Conn.WriteJSON(data); err != nil {
			log.Printf("ERROR [%v] writing to connection: %v\n", p.SessionId, err)
		}
	}
}
