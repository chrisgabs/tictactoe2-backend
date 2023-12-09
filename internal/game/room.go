package game

import "log"

type Room struct {
	RoomId      string
	Player1     *Player
	Player2     *Player
	Board       Board
	Receiver    chan *MessageData
	GameOngoing bool
	Game        *GameInstance
}

// StartGame Start receiving MessageData from players in the room
func (r *Room) StartGame() {
	log.Printf("Room [%v] game starting with players [%v] and [%v]\n", r.RoomId, r.Player1.SessionId, r.Player2.SessionId)
	r.GameOngoing = true
	// implement different types of message types.
	// ex: if leave, restart the game, force other player to refresh board
	defer func() {
		log.Printf("Room [%v] Stopping: Insufficient number of players", r.RoomId)
		r.GameOngoing = false
	}()

	// PlayerNumber
	// 1 - send to 2
	// 2 - send to 1
	// 0 - send to all
	for data := range r.Receiver {
		if r.Player1 == nil || r.Player2 == nil {
			if data.EventType != Leave {
				return
			}
		}
		if data.PlayerNumber == 1 {
			r.Player2.ReceiveChannel <- data
		} else if data.PlayerNumber == 2 {
			r.Player1.ReceiveChannel <- data
		} else {
			r.Player1.ReceiveChannel <- data
			r.Player2.ReceiveChannel <- data
		}
	}
}

// AddPlayer check whether to put player in p1 or p2
// notifies opponent that player has joined via WS
// set's p's playerNumber variable
func (r *Room) AddPlayer(p *Player) (int, bool) {
	playerNumber := 0
	if r.Player1 == nil {
		r.Player1 = p
		playerNumber = 1
	} else if r.Player2 == nil {
		r.Player2 = p
		playerNumber = 2
	} else {
		// return error
		return 0, false
	}
	p.PlayerNumber = playerNumber
	p.Room = r
	if r.Player1 != nil && r.Player2 != nil {
		go r.StartGame()
		r.notifyOpponent(p)
	}
	return playerNumber, true
}

func (r *Room) RemovePlayer(p *Player) {
	if p.PlayerNumber == 1 {
		r.Player1 = nil
	} else {
		r.Player2 = nil
	}
	if r.Player1 == nil && r.Player2 == nil {
		delete(r.Game.Rooms, r.RoomId)
		log.Printf("Killing room [%v]: No more players", r.RoomId)
	}
}

func (r *Room) notifyOpponent(joiningPlayer *Player) {
	data := make(map[string]string)
	data["opponentDisplayName"] = joiningPlayer.DisplayName
	notification := MessageData{EventType: Join, PlayerNumber: joiningPlayer.PlayerNumber, Data: data}
	//if r.GameOngoing {
	r.Receiver <- &notification // gets stuck here if notification is not consumed (when opponent does not exist or game is not yet started)
	//}
}
