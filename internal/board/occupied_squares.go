package board

import "github.com/Glenn444/golang-chess/utils"

/*
- Check if
*/
func Occupied_squares(g GameState, pos string)(string)  {
	row,col := utils.Chess_notation_to_indices(pos)
	square := g.Board[row][col]
	if !square.Occupied{
		return  "EMPTY"
	}else if square.Piece.GetColor() == g.CurrentPlayer{
		return  "OWN_PIECE"
	}
	return "OPPONENT_PIECE"
}