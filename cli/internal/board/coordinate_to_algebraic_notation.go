package board

import (
	"fmt"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/utils"
)

func CoordinateToAlgebraic(game pieces.GameState, move string) string {
	//move - e2e4 should be e4 if pawn and Qe4 if queen e.t.c
	//ensure move is not a capture,pwan move
	//check if to position is occupied
	castle, move := isCastlingMove(move)
	if castle {
		fmt.Printf("it's a castling move")
		return ""
	}
	
	from := move[0:2]
	to := move[2:4]
	rowto, colto := utils.Chess_notation_to_indices(to)
	occupied := game.Board[rowto][colto].Occupied
	if occupied {
		rowfrom, colfrom := utils.Chess_notation_to_indices(from)
		pieceType := game.Board[rowfrom][colfrom].Piece.GetPieceType()

		if pieceType != "P" {
			return fmt.Sprintf("%sx%s", pieceType, to)
		}
		return fmt.Sprintf("%sx%s", string(from[0]), to)
	} else {

		rowfrom, colfrom := utils.Chess_notation_to_indices(from)
		pieceType := game.Board[rowfrom][colfrom].Piece.GetPieceType()

		if pieceType != "P" {
			return fmt.Sprintf("%s%s", pieceType, to)
		}
		return to
	}

}

func isCastlingMove(move string) (bool, string) {
	castlingMoves := map[string]string{
		"e1g1": "O-O",   // white kingside
		"e1c1": "O-O-O", // white queenside
		"e8g8": "O-O",   // black kingside
		"e8c8": "O-O-O", // black queenside
	}
	if notation, ok := castlingMoves[move]; ok {
		return true, notation
	}
	return false, ""
}
