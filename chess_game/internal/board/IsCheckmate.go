package board

import (
	"slices"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/internal/utils/chess"
)

// IsCheckmate returns true when the current player is in check and has no
// legal move that escapes check.
func IsCheckmate(game pieces.GameState) bool {
	if !IsKinginCheck(game) {
		return false
	}
	return !currentPlayerHasLegalMove(game)
}

// currentPlayerHasLegalMove tries every piece's every candidate square and
// returns true as soon as it finds a move that does not leave the king in check.
func currentPlayerHasLegalMove(game pieces.GameState) bool {
	for _, row := range game.Board {
		for _, square := range row {
			if !square.Occupied || square.Piece.GetColor() != game.CurrentPlayer {
				continue
			}
			sourcePos := square.Piece.GetPosition()
			for _, targetPos := range square.Piece.GetLegalSquares(game) {
				if targetPos == sourcePos {
					continue
				}
				if !moveLeavesKingInCheck(game, sourcePos, targetPos) {
					return true
				}
			}
		}
	}
	return false
}

// moveLeavesKingInCheck simulates the move from→to on a board copy and
// returns true if the current player's king is still in check afterwards.
func moveLeavesKingInCheck(game pieces.GameState, from, to string) bool {
	boardCopy := Create_board()
	CopyBoard(boardCopy, game.Board)

	fromRow, fromCol, _ := chess.ChessNotationToIndices(from)
	toRow, toCol, _ := chess.ChessNotationToIndices(to)

	piece := boardCopy[fromRow][fromCol].Piece
	if piece == nil {
		return true
	}

	boardCopy[fromRow][fromCol] = pieces.Square{Occupied: false, Piece: nil}
	pieceCopy := piece.Clone()
	pieceCopy.AssignPosition(to)
	boardCopy[toRow][toCol] = pieces.Square{Occupied: true, Piece: pieceCopy}

	// Locate the current player's king on the copied board.
	var kingPos string
	for _, row := range boardCopy {
		for _, sq := range row {
			if sq.Occupied && sq.Piece.GetPieceType() == "K" && sq.Piece.GetColor() == game.CurrentPlayer {
				kingPos = sq.Piece.GetPosition()
			}
		}
	}
	if kingPos == "" {
		return true
	}

	opponentColor := "b"
	if game.CurrentPlayer == "b" {
		opponentColor = "w"
	}

	tempGame := pieces.GameState{Board: boardCopy}
	for _, row := range boardCopy {
		for _, sq := range row {
			if sq.Occupied && sq.Piece.GetColor() == opponentColor {
				if slices.Contains(sq.Piece.GetLegalSquares(tempGame), kingPos) {
						return true
					}
			}
		}
	}
	return false
}
