package board

import (
	"slices"

	"github.com/Glenn444/golang-chess/internal/pieces"
)


func IsKinginCheck(game pieces.GameState)bool{
	var kingPos string
	for _,squares := range game.Board{
		for _, square := range squares{
			if square.Occupied && square.Piece.GetPieceType() == "K" && square.Piece.GetColor() == game.CurrentPlayer{
				kingPos = square.Piece.GetPosition()
			}
		}
	}
	for _,squares := range game.Board{
		for _,square := range squares{
			if square.Occupied && square.Piece.GetColor() != game.CurrentPlayer{
				legalSquares := square.Piece.GetLegalSquares(game)

				if slices.Contains(legalSquares,kingPos){
					return true
				}
			}
		}
	}
	return false
}