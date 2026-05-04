package board

import (
	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/internal/utils/chess"
)

/*
- Returns occupied squares
*/
func Occupied_squares(g pieces.GameState, pos string)(bool,string)  {
	destrow,destcol, err := chess.ChessNotationToIndices(pos)
	if err != nil {
		return false, ""
	}
	square := g.Board[destrow][destcol]
	if !square.Occupied{
		return  false,"EMPTY"
	}else if square.Piece.GetColor() == g.CurrentPlayer{
		return  true, "OWN_PIECE"
	}
	return true, "OPPONENT_PIECE"
}

