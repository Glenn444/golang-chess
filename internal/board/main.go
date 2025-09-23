package board

import (
	"github.com/Glenn444/golang-chess/internal/pieces"
)

type Square struct {
	Occupied bool
	Piece    pieces.PieceInterface
}

type GameState struct {
	CurrentPlayer string
	Board         [][]Square
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

func CurrentPlayer_Occupied_Piece_position(g GameState, pos string) string {
	pieceType := string(pos[0])
	pos_sub := pos[1:]
	for _, square := range g.Board {
		for _, s := range square {
			//fmt.Print(s)
			//fmt.Print(s.Piece.GetPieceType() == pieceType)
			if s.Occupied && s.Piece.GetColor() == g.CurrentPlayer && s.Piece.GetPieceType() == pieceType {
				legal_squares := s.Piece.GetLegalSquares()
				//fmt.Printf("legal squares: %v\n", legal_squares)
				for _, c_pos := range legal_squares {
					//fmt.Printf("c_pos: %s, pos: %s\n",c_pos,pos_sub)
					if c_pos == pos_sub {
						return s.Piece.GetPosition()
					}
				}
			} else if s.Occupied && s.Piece.GetColor() == g.CurrentPlayer {
				legal_squares := s.Piece.GetLegalSquares()
				//fmt.Printf("legal squares: %v\n",legal_squares)
				for _, c_pos := range legal_squares {
					if c_pos == pos {
						return s.Piece.GetPosition()
					}
				}
			}
		}
	}
	return ""
}
