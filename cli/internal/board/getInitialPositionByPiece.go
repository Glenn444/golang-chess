package board

import (
	"slices"
	"errors"

	"github.com/Glenn444/golang-chess/internal/pieces"
)

func GetInitialPositionByPiece(destinationPos string, pieceType string, gm pieces.GameState)(piece pieces.PieceInterface,err error){
	for _,squares := range gm.Board{
		for _, square := range squares{
			if square.Piece.GetPieceType() == pieceType && square.Piece.GetColor() == gm.CurrentPlayer{
				if slices.Contains(square.Piece.GetLegalSquares(gm), destinationPos) {
						return square.Piece, nil
					}
			}
		}
	}
	return nil, errors.New("no piece found")
}