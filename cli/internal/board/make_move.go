package board

import (
	"fmt"

	"github.com/Glenn444/golang-chess/utils"
)

func Move(game *GameState, move string) error {
	
	move_type := string(move[2])
	//pawn move
	if move_type != "x" {
		move_pos := string(move[1:])
		if len(move) == 2 {
			move_pos = move
		}

		sourcepos, err := CurrentPlayer_Occupied_Piece_position(*game, move)
		if err != nil {
			return err
		}

		destrow, destcol := utils.Chess_notation_to_indices(move_pos)
		sourcerow, sourcecol := utils.Chess_notation_to_indices(sourcepos)

		piece := game.Board[sourcerow][sourcecol].Piece
		piece.AssignPosition(move_pos)
		//fmt.Printf("Piece pos: %s\n",piece.GetPosition())

		//clear the source square
		game.Board[sourcerow][sourcecol] = Square{
			Occupied: false,
			Piece:    nil,
		}

		//destination square
		game.Board[destrow][destcol] = Square{
			Occupied: true,
			Piece:    piece,
		}

		//squareOccupied, val := Occupied_squares(*game, move_pos)
		occupiedPositions := GetAllOccupiedSquares(*game)
		fmt.Printf("%v occupied squares: %v \n", game.CurrentPlayer, occupiedPositions)
		// if squareOccupied {
		// 	fmt.Printf("%v %s\n",squareOccupied,val)

		// }
		//change current player after making move
		if game.CurrentPlayer == "w" {
			game.CurrentPlayer = "b"
		} else {
			game.CurrentPlayer = "w"
		}

		return nil
	}else{
		destCapturePos := string(move[2:])
		piceType := string(move[0])

		boardFile := map[string]bool{
			"a":true,"b":true,"c":true,"d":true,"e":true,"f":true,"g":true,"h":true,
		}
		

		//1. determine if it a pawn capture or another piece
		if boardFile[piceType]{
			
		}

	}
}
