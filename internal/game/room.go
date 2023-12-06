package game

import "log"

type Room struct {
	RoomId      string
	Player1     *Player
	Player2     *Player
	Board       Board
	Receiver    chan *MessageData
	GameOngoing bool
}

// StartGame Start receiving MessageData from players in the room
func (r *Room) StartGame() {
	log.Printf("Room [%v] game starting with players [%v] and [%v]\n", r.RoomId, r.Player1.SessionId, r.Player2.SessionId)
	// implement different types of message types.
	// ex: if leave, restart the game, force other player to refresh board
	defer log.Printf("Room [%v] Stopping: Insufficient number of players", r.RoomId)
	for data := range r.Receiver {
		if r.Player1 == nil || r.Player2 == nil {
			return
		}
		if data.PlayerNumber == 1 {
			r.Player2.ReceiveChannel <- data
		} else {
			r.Player1.ReceiveChannel <- data
		}
	}
}

// AddPlayer check whether to put player in p1 or p2
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
	}
	return playerNumber, true
}
