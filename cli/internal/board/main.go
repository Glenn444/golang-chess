package board

import (
	"errors"
	"slices"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/utils"
)

type Square struct {
	Occupied bool
	Piece    pieces.PieceInterface
}

// type GameState struct {
// 	CurrentPlayer string
// 	Board         [][]Square
// }

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

func CurrentPlayer_Occupied_Piece_position(g pieces.GameState, pos string) (string,error) {
	
	occupied_squares := GetAllOccupiedSquares(g)

	// Check if it's a pawn move (no piece prefix)
	if len(pos) == 2 && pos[0] >= 'a' && pos[0] <= 'h' {
		// This is a pawn move like "d4", "e5", etc.
		pieceType := "P" // or whatever you use for pawns
		destpos := pos

		for _, squares := range g.Board {
			for _, s := range squares {
				if s.Occupied && s.Piece.GetColor() == g.CurrentPlayer && s.Piece.GetPieceType() == pieceType {
					pieces_squares := s.Piece.GetLegalSquares(g)
					legal_squares := utils.RemoveOwnOccupiedSquares(pieces_squares, occupied_squares)
					if slices.Contains(legal_squares, destpos) {
							return s.Piece.GetPosition(),nil
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

					pieces_squares := s.Piece.GetLegalSquares(g)
					legal_squares := utils.RemoveOwnOccupiedSquares(pieces_squares, occupied_squares)
					//fmt.Printf("legal squares: %v\n", legal_squares)

					if slices.Contains(legal_squares, destpos_sub) {
							return s.Piece.GetPosition(),nil
						}
				}
			}
		}
	}

	return "", errors.New("invalid move\n")
}
