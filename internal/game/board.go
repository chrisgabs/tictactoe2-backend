package game

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

type Board struct {
	b [3][3]pieceData
}

type pieceData struct {
	playerNumber int
	piece        string
}

func (board *Board) placePiece(dropData DropData) {
	if !dropData.IsValidMove {
		return
	}
	cell := dropData.Cell
	cellNum, _ := strconv.Atoi(cell)

	x := cellNum / 3
	y := cellNum % 3
	board.b[x][y] = pieceData{
		playerNumber: dropData.PlayerNumber,
		piece:        dropData.Piece,
	}
}

// AsRawMessage transfrom board data to make it readable by frontend
func (board *Board) AsRawMessage() json.RawMessage {
	idx := 0
	var rawMessage = make(map[string]string)
	for i := 0; i < len(board.b); i++ {
		for j := 0; j < len(board.b[i]); j++ {
			p := board.b[i][j]
			rawMessage[strconv.Itoa(idx)] = fmt.Sprintf("%v,%v", p.playerNumber, p.piece)
			idx += 1
		}
	}
	dataJSON, err := json.Marshal(rawMessage)
	if err != nil {
		log.Printf("ERROR AsRawMessage: %v\n", err)
		return nil
	}
	return dataJSON
}

func (board *Board) checkForWin() bool {
	b := board.b
	// Check rows and columns
	for i := 0; i < 3; i++ {
		if b[i][0].playerNumber != 0 { // could be incorporated inside inner condition similar to diagonal condition
			if b[i][0].playerNumber == b[i][1].playerNumber && b[i][1].playerNumber == b[i][2].playerNumber {
				return true
			}
		}
		if b[0][i].playerNumber != 0 {
			if b[0][i].playerNumber == b[1][i].playerNumber && b[1][i].playerNumber == b[2][i].playerNumber {
				return true
			}
		}
	}

	if ((b[0][0].playerNumber == b[1][1].playerNumber && b[1][1].playerNumber == b[2][2].playerNumber) && b[0][0].playerNumber != 0) ||
		((b[0][2].playerNumber == b[1][1].playerNumber && b[1][1].playerNumber == b[2][0].playerNumber) && b[0][2].playerNumber != 0) {
		return true
	}
	return false
}

func (board *Board) ResetBoard() {
	board.b = [3][3]pieceData{}
}

func CreateEmptyBoard() Board {
	return Board{
		b: [3][3]pieceData{},
	}
}
