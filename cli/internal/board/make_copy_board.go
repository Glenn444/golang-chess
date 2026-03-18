package board

import "github.com/Glenn444/golang-chess/internal/pieces"




func CopyBoard(copyA[][]pieces.Square,copyB[][]pieces.Square){
	//copy b to A
	for i,squares := range copyB{
		for j,square := range squares{
			copyA[i][j] = pieces.Square{
				Occupied: square.Occupied,
				Piece: square.Piece.Clone(),
			}
		}
	}
}