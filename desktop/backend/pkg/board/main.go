package board

import (


	"github.com/Glenn444/golang-chess/backend/pkg/pieces"
	"github.com/Glenn444/golang-chess/backend/utils"
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
	occupied_squares := GetAllOccupiedSquares(g)

	// Check if it's a pawn move (no piece prefix)
	if len(pos) == 2 && pos[0] >= 'a' && pos[0] <= 'h' {
		// This is a pawn move like "d4", "e5", etc.
		pieceType := "P" // or whatever you use for pawns
		destpos := pos

		for _, square := range g.Board {
			for _, s := range square {
				if s.Occupied && s.Piece.GetColor() == g.CurrentPlayer && s.Piece.GetPieceType() == pieceType {
					pieces_squares := s.Piece.GetLegalSquares()
					legal_squares := utils.RemoveOwnOccupiedSquares(pieces_squares, occupied_squares)
					for _, c_pos := range legal_squares {
						if c_pos == destpos {
							return s.Piece.GetPosition()
						}
					}
				}
			}
		}
	} else {
		// This is a piece move like "Nc3", "Qd4", etc.
		pieceType := string(pos[0])
		destpos_sub := pos[1:]
		for _, square := range g.Board {
		for _, s := range square {
			//fmt.Print(s)
			//fmt.Print(s.Piece.GetPieceType() == pieceType)
			if s.Occupied && s.Piece.GetColor() == g.CurrentPlayer && s.Piece.GetPieceType() == pieceType {

				pieces_squares := s.Piece.GetLegalSquares()
				legal_squares := utils.RemoveOwnOccupiedSquares(pieces_squares,occupied_squares)
				//fmt.Printf("legal squares: %v\n", legal_squares)
				for _, c_pos := range legal_squares {
					//fmt.Printf("c_pos: %s, pos: %s\n",c_pos,pos_sub)
					if c_pos == destpos_sub {
						return s.Piece.GetPosition()
					}
				}
			}
		}
	}}

	return ""
	}
