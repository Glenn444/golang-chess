package board

import "github.com/Glenn444/golang-chess/utils"

func CapturePiece(g *GameState,move string) error {
	destCapturePos := string(move[2:])
	piceType := string(move[0])

	boardFile := map[string]bool{
			"a":true,"b":true,"c":true,"d":true,"e":true,"f":true,"g":true,"h":true,
	}

	row, col := utils.Chess_notation_to_indices(destCapturePos)
	pieceSquare := g.Board[row][col]

	if pieceSquare.Occupied && g.CurrentPlayer != pieceSquare.Piece.GetColor(){
		//valid capture
		if boardFile[piceType]{
			//this is a pawn capture
			
		}
		pieceSquare.Piece = 
	}
}