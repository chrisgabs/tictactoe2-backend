package game

import (
	"github.com/chrisgabs/tictactoe2-backend/internal/util"
	"github.com/gorilla/websocket"
	"log"
)

type GameInstance struct {
	Rooms   map[string]*Room
	Players map[string]*Player
}

// CreatePlayer initializes a player then adds it to list of players
func (g *GameInstance) CreatePlayer(addr string, displayName string, ws *websocket.Conn) (*Player, bool) {
	if _, exists := g.Players[addr]; exists {
		log.Printf("Player already in list of players: %v\n", addr)
		return nil, false
	} else {
		g.Players[addr] = &Player{
			SessionId:      addr,
			PlayerNumber:   0,
			Room:           &Room{},
			DisplayName:    displayName,
			ReceiveChannel: make(chan *MessageData),
			Conn:           ws,
			WSConnected:    false, // set to true when actual ws connection is confirmed
			Game:           g,
		}
		return g.Players[addr], true
	}
}

func (g *GameInstance) CreateRoom() *Room {
	var name = ""
	for {
		name = "r_" + util.RandomString(3)
		if _, exists := g.Rooms[name]; !exists {
			break
		}
	}
	g.Rooms[name] = &Room{
		RoomId:      name,
		Board:       CreateEmptyBoard(),
		Receiver:    make(chan *MessageData),
		GameOngoing: false,
	}
	return g.Rooms[name]
}

func (g *GameInstance) AddPlayerToRoom(p *Player, r *Room) (int, bool) {
	prevPlayerNum := p.PlayerNumber
	prevRoom := p.Room
	playerNum, playerAddSuccess := r.AddPlayer(p)
	if playerAddSuccess {
		// Check if not nil room
		// Remove player in previous room
		// Delete the room if nobody inside
		if prevRoom.RoomId != "" {
			if prevPlayerNum == 1 {
				prevRoom.Player1 = nil
			} else {
				prevRoom.Player2 = nil
			}
			g.safeDeleteRoom(prevRoom)
		}
		return playerNum, true
	} else {
		return 0, playerAddSuccess
	}
}

func (g *GameInstance) safeDeleteRoom(r *Room) {
	if r.Player1 == nil && r.Player2 == nil {
		delete(g.Rooms, r.RoomId)
	}
}
