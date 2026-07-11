package board

import "github.com/Glenn444/golang-chess/internal/pieces"

func CopyBoard(destCopy [][]pieces.Square, sourceCopy [][]pieces.Square) {
	//copy b to A
	for i, squares := range sourceCopy {
		for j, square := range squares {
			if square.Occupied {
				destCopy[i][j] = pieces.Square{
					Occupied: square.Occupied,
					Piece:    square.Piece.Clone(),
				}
			} else {
				destCopy[i][j] = pieces.Square{
					Occupied: square.Occupied,
					Piece:    nil,
				}
			}

		}
	}
}
