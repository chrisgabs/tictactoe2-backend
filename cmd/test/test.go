package main

type Board struct {
	b [3][3]pieceData
}

type pieceData struct {
	playerNumber int
	piece        string
}

func checkForWin(b *Board) bool {
	// Check rows and columns
	for i := 0; i < 3; i++ {
		if (b.b[i][0].playerNumber == b.b[i][1].playerNumber && b.b[i][1].playerNumber == b.b[i][2].playerNumber) ||
			(b.b[0][i].playerNumber == b.b[1][i].playerNumber && b.b[1][i].playerNumber == b.b[2][i].playerNumber) {
			return true
		}
	}

	// Check diagonals
	if (b.b[0][0].playerNumber == b.b[1][1].playerNumber && b.b[1][1].playerNumber == b.b[2][2].playerNumber) ||
		(b.b[0][2].playerNumber == b.b[1][1].playerNumber && b.b[1][1].playerNumber == b.b[2][0].playerNumber) {
		return true
	}

	return false
}

func main() {

}
