package board

import (
	"fmt"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/utils"
)


func CoordinateToAlgebraic(game pieces.GameState, move string)string{
	//move - e2e4 should be e4 if pawn and Qe4 if queen e.t.c
	//ensure move is not a capture,pwan move
	from := move[0:2]
	to := move[2:4]

	row,col := utils.Chess_notation_to_indices(from)
	pieceType := game.Board[row][col].Piece.GetPieceType()

	if pieceType != "P"{
		return fmt.Sprintf("%s%s",pieceType,to)
	}
	return to

}


