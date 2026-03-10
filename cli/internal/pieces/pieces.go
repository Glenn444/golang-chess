package pieces

import (
	"fmt"

)

type Square struct {
	Occupied bool
	Piece    PieceInterface
}

type GameState struct {
	CurrentPlayer string
	Board         [][]Square
}

type PieceInterface interface{
	GetLegalSquares(g GameState) []string
	GetColor() string
    GetPosition() string
    GetPieceType() string
	AssignPosition(pos string)
	String() string
}
func PrintBoard(initialBoard_position [][]Square){

	fmt.Printf("      a  b  c  d  e  f  g  h\n")
	for i, row := range initialBoard_position {
		
		fmt.Printf("%d", i+1)
		fmt.Printf("    ")
		for _, s := range row {
			
			if s.Occupied{
			fmt.Printf("%v", s.Piece.String())
			}else{
				fmt.Printf("[ ]")
			}
			
		}
	fmt.Printf("\n")

	}
	fmt.Printf("      a  b  c  d  e  f  g  h\n")
}


func Create_board() [][]Square {

	rows, cols := 8, 8

	board := make([][]Square, rows)

	for i := range board {
		board[i] = make([]Square, cols)
	}

	// First, initialize all squares as empty
	for i := range board {
		for j := range board[i] {
			board[i][j] = Square{
				Occupied: false,
				Piece:    nil,
			}
		}
	}

	return board
}

