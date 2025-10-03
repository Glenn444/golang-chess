package board

import (
	"fmt"

	"github.com/Glenn444/golang-chess/backend/utils"
)

func Move(game *GameState, move string) (g *GameState, err error){
	var move_pos string
	sourcepos := CurrentPlayer_Occupied_Piece_position(*game, move)

	move_pos = string(move[1:])
	//pawn move
	if len(move) == 2{
		move_pos = move
	}
	
	destrow, destcol := utils.Chess_notation_to_indices(move_pos)
	sourcerow, sourcecol := utils.Chess_notation_to_indices(sourcepos)

	piece := game.Board[sourcerow][sourcecol].Piece
	piece.AssignPosition(move_pos)
	fmt.Printf("Piece pos: %s\n",piece.GetPosition())

	//clear the source square
	game.Board[sourcerow][sourcecol] = Square{
		Occupied: false,
		Piece:    nil,
	}

	squareOccupied, val := Occupied_squares(*game, move_pos)
	if squareOccupied {
		fmt.Printf("%v %s\n",squareOccupied,val)
		return nil,fmt.Errorf("square occupied")
	}

	//destination square
	game.Board[destrow][destcol] = Square{
		Occupied: true,
		Piece:    piece,
	}

	//change current player after making move
	if game.CurrentPlayer == "w" {
		game.CurrentPlayer = "b"
	} else {
		game.CurrentPlayer = "w"
	}
	return  game,nil
}
