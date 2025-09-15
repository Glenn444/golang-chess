package board

import "github.com/Glenn444/golang-chess/internal/pieces"



type Square struct{
Occupied bool
Piece pieces.PieceInterface
}


func Create_board() [][]Square  {
	// n := Square{
	// 	Occupied: true,
	// 	Piece: piece{
	// 		Color: "white",
	// 		PieceType: "N",
	// 		Position: "b1",
	// 	},
	// }
	
	rows,cols := 8,8

	board := make([][]Square,rows)

	for i := range board{
		board[i] = make([]Square, cols)
	}

	return board
}