package board

import "github.com/Glenn444/golang-chess/internal/pieces"



type Square struct{
Occupied bool
Piece pieces.PieceInterface
}

type GameState struct {
	CurrentPlayer string
	Board [][]Square
}

func Create_board() [][]Square  {
	
	rows,cols := 8,8

	board := make([][]Square,rows)

	for i := range board{
		board[i] = make([]Square, cols)
	}

	return board
}