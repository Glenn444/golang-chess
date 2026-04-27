package board

import (
	"errors"
	"slices"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/internal/utils/chess"
)

type Square struct {
	Occupied bool
	Piece    pieces.PieceInterface
}

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
					legal_squares := chess.RemoveOwnOccupiedSquares(pieces_squares, occupied_squares)
					if slices.Contains(legal_squares, destpos) {
							return s.Piece.GetPosition(),nil
						}
				}
			}
		}
	} else if len(pos) == 4 {
		// This is a piece move like "Nac3", "Rbc3", etc.
		pieceType := string(pos[0])
		destpos_sub := pos[2:]
		destCol := string(pos[1])
		for _, square := range g.Board {
			for _, s := range square {
				
				if s.Occupied && s.Piece.GetColor() == g.CurrentPlayer && s.Piece.GetPieceType() == pieceType {

					pieces_squares := s.Piece.GetLegalSquares(g)
					legal_squares := chess.RemoveOwnOccupiedSquares(pieces_squares, occupied_squares)

					piecePos := s.Piece.GetPosition() //a2,a3,b5
					if slices.Contains(legal_squares, destpos_sub) && string(piecePos[0]) == destCol {
							return s.Piece.GetPosition(),nil
						}
				}
			}
		}
		return "",errors.New("invalid move\n")

	}else {
		// This is a piece move like "Nc3", "Qd4", etc.
		pieceType := string(pos[0])
		destpos_sub := pos[1:]
		for _, square := range g.Board {
			for _, s := range square {
				
				if s.Occupied && s.Piece.GetColor() == g.CurrentPlayer && s.Piece.GetPieceType() == pieceType {

					pieces_squares := s.Piece.GetLegalSquares(g)
					legal_squares := chess.RemoveOwnOccupiedSquares(pieces_squares, occupied_squares)
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
