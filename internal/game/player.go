package game

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type Player struct {
	SessionId      string `json:"sessionId"`
	PlayerNumber   int    `json:"playerNumber"`
	DisplayName    string `json:"displayName"`
	Room           *Room
	ReceiveChannel chan *MessageData
	Conn           *websocket.Conn
	WSConnected    bool
	Game           *GameInstance
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

func (p *Player) startVultureTask() {
	timer := time.NewTimer(5 * time.Second)
	timedOut := false
	go func() {
		<-timer.C
		timedOut = true
	}()
	for {
		if p.WSConnected {
			log.Printf("Player [%v] successfully reconnected", p.DisplayName)
			return
		}
		if timedOut {
			log.Printf("Player [%v] TIMEDOUT", p.DisplayName)
			close(p.ReceiveChannel)
			p.Room.RemovePlayer(p)
			delete(p.Game.Players, p.SessionId)
			return
		}
	}
}

func (p *Player) StartListeningToClient() {

	defer func(Conn *websocket.Conn) {
		err := Conn.Close()
		if err != nil {
			log.Printf("Error closing WS connection of [%v]%v : %v\n", p.SessionId, p.DisplayName, err)
		}
		log.Printf("Closed WS connection of [%v]%v\n", p.SessionId, p.DisplayName)
		p.WSConnected = false
		go p.startVultureTask()
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

		if message.EventType == Connect {
			data := make(map[string]interface{})
			data["playerNumber"] = p.PlayerNumber
			data["roomId"] = p.Room.RoomId
			data["boardData"] = p.Room.Board.AsRawMessage()
			response := MessageData{Connect, p.PlayerNumber, data}
			if err := p.Conn.WriteJSON(response); err != nil {
				log.Printf("error in writing JSON 64: %v\n", err)
			} else {
				p.WSConnected = true
			}
		}

		if p.Room.GameOngoing || message.EventType == Leave {
			// check what to do based on message type
			switch eType := message.EventType; eType {
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
			case Reset:
				data := make(map[string]string)
				data["DisplayName"] = p.DisplayName
				response := MessageData{Reset, 0, data}
				p.Room.Receiver <- &response
				p.Room.Board.ResetBoard()
			case Leave:
				data := make(map[string]string)
				data["DisplayName"] = p.DisplayName
				response := MessageData{Leave, p.PlayerNumber, data}
				p.Room.Receiver <- &response
				p.Room.RemovePlayer(p)
				p.Room.Board.ResetBoard()
			} // switch
		} // if
	} // for

} // function

// StartListeningToRoom to be called only by the room player is in to ensure that p.ReceiveChannel exists
func (p *Player) StartListeningToRoom() {
	log.Printf("Player [%v] started listening to room [%v]\n", p.DisplayName, p.Room.RoomId)
	defer func() {
		log.Printf("[%v] Stopped listening to room [%v]\n", p.SessionId, p.Room.RoomId)
	}()
	for data := range p.ReceiveChannel {
		if err := p.Conn.WriteJSON(data); err != nil {
			log.Printf("ERROR [%v] writing to connection: %v\n", p.SessionId, err)
			return
		}
	}
}
