package board

import (
	"github.com/Glenn444/golang-chess/internal/pieces"
)

func Create_board() [][]pieces.Square {

	rows, cols := 8, 8

	board := make([][]pieces.Square, rows)

	for i := range board {
		board[i] = make([]pieces.Square, cols)
	}

	// First, initialize all squares as empty
	for i := range board {
		for j := range board[i] {
			board[i][j] = pieces.Square{
				Occupied: false,
				Piece:    nil,
			}
		}
	}

	return board
}
