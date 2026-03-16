package board

import (
	"fmt"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/utils"
)

func Move(game *pieces.GameState, move string) error {
	var move_pos string
	move_pos = string(move[1:])
	moveType := string(move[1])

	if moveType == "x" || moveType == "X"{
		fmt.Printf("this is a capture move\n")
		err := CapturePiece(game,move)
		if err != nil{
			return err
		}
		return nil

	}

	
	sourcepos,err := CurrentPlayer_Occupied_Piece_position(*game, move)
	if err != nil{
		return err
	}
	//fmt.Printf("sourcepos: %v",sourcepos)
	

	//pawn move
	if len(move) == 2{
		move_pos = move
	}
	
	destrow, destcol := utils.Chess_notation_to_indices(move_pos)
	sourcerow, sourcecol := utils.Chess_notation_to_indices(sourcepos)

	piece := game.Board[sourcerow][sourcecol].Piece
	piece.AssignPosition(move_pos)
	//fmt.Printf("Piece pos: %s\n",piece.GetPosition())

	//clear the source square
	game.Board[sourcerow][sourcecol] = pieces.Square{
		Occupied: false,
		Piece:    nil,
	}

	
	//destination square
	game.Board[destrow][destcol] = pieces.Square{
		Occupied: true,
		Piece:    piece,
	}

	//squareOccupied, val := Occupied_squares(*game, move_pos)
	occupiedPositions := GetAllOccupiedSquares(*game)
	fmt.Printf("%v occupied squares: %v \n",game.CurrentPlayer,occupiedPositions)
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
}
