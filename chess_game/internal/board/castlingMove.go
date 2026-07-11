package board

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Glenn444/golang-chess/internal/pieces"
)

// CastlingMove performs kingside ("O-O", "e1g1", "e8g8") or queenside
// ("O-O-O", "e1c1", "e8c8") castling for the current player.
func CastlingMove(gameState *pieces.GameState, move string) error {
	row := 0
	if gameState.CurrentPlayer == "b" {
		row = 7
	}

	var kingside bool
	switch move {
	case "O-O", "e1g1", "e8g8":
		kingside = true
	case "O-O-O", "e1c1", "e8c8":
		kingside = false
	default:
		return errors.New("invalid castling move")
	}
	// A coordinate move must belong to the player on move.
	if (strings.HasPrefix(move, "e1") && gameState.CurrentPlayer != "w") ||
		(strings.HasPrefix(move, "e8") && gameState.CurrentPlayer != "b") {
		return errors.New("invalid castling move")
	}

	rank := row + 1
	sq := func(col int) string { return fmt.Sprintf("%c%d", 'a'+col, rank) }

	var rookFromCol, kingToCol, rookToCol int
	var emptyCols []int
	var kingPath []string // squares the king stands on or crosses; none may be attacked
	if kingside {
		rookFromCol, kingToCol, rookToCol = 7, 6, 5
		emptyCols = []int{5, 6}
		kingPath = []string{sq(4), sq(5), sq(6)}
	} else {
		rookFromCol, kingToCol, rookToCol = 0, 2, 3
		emptyCols = []int{1, 2, 3}
		kingPath = []string{sq(4), sq(3), sq(2)}
	}

	c := &gameState.Castle
	var kingMoved, rookMoved *bool
	if gameState.CurrentPlayer == "w" {
		kingMoved = &c.WhiteKingMoved
		if kingside {
			rookMoved = &c.WhiteRookKingsideMoved
		} else {
			rookMoved = &c.WhiteRookQueensideMoved
		}
	} else {
		kingMoved = &c.BlackKingMoved
		if kingside {
			rookMoved = &c.BlackRookKingsideMoved
		} else {
			rookMoved = &c.BlackRookQueensideMoved
		}
	}

	if *kingMoved || *rookMoved {
		return errors.New("castling not allowed: king or rook has already moved")
	}
	kingSq := gameState.Board[row][4]
	rookSq := gameState.Board[row][rookFromCol]
	if !kingSq.Occupied || kingSq.Piece.GetPieceType() != "K" || kingSq.Piece.GetColor() != gameState.CurrentPlayer ||
		!rookSq.Occupied || rookSq.Piece.GetPieceType() != "R" || rookSq.Piece.GetColor() != gameState.CurrentPlayer {
		return errors.New("castling not allowed: king or rook not on its starting square")
	}
	for _, col := range emptyCols {
		if gameState.Board[row][col].Occupied {
			return errors.New("castling not allowed: squares between king and rook are occupied")
		}
	}
	if CastlingSquareisAttacked(gameState, kingPath) {
		return errors.New("castling not allowed: king is in check or would pass through an attacked square")
	}

	king := kingSq.Piece
	rook := rookSq.Piece
	king.AssignPosition(sq(kingToCol))
	rook.AssignPosition(sq(rookToCol))
	gameState.Board[row][4] = pieces.Square{}
	gameState.Board[row][rookFromCol] = pieces.Square{}
	gameState.Board[row][kingToCol] = pieces.Square{Occupied: true, Piece: king}
	gameState.Board[row][rookToCol] = pieces.Square{Occupied: true, Piece: rook}

	*kingMoved = true
	*rookMoved = true
	gameState.EnPassantTarget = ""
	gameState.StockfishGame = append(gameState.StockfishGame, sq(4)+sq(kingToCol))

	switchPlayer(gameState)
	return nil
}

// CastlingSquareisAttacked reports whether an opponent piece attacks any of
// the given squares.
func CastlingSquareisAttacked(game *pieces.GameState, castlingSquares []string) bool {
	for _, squares := range game.Board {
		for _, square := range squares {
			if square.Occupied && square.Piece.GetColor() != game.CurrentPlayer {
				legalSquares := square.Piece.GetLegalSquares(game)
				for _, castlingSquare := range castlingSquares {
					if slices.Contains(legalSquares, castlingSquare) {
						return true
					}
				}
			}
		}
	}
	return false
}
