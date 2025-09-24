package board

import (
	"fmt"

	"github.com/Glenn444/golang-chess/utils"
)

func Move(game *GameState, move string) {

	sourcepos := CurrentPlayer_Occupied_Piece_position(*game, move)

	move_pos := string(move[1:])
	destrow, destcol := utils.Chess_notation_to_indices(move_pos)
	sourcerow, sourcecol := utils.Chess_notation_to_indices(sourcepos)

	piece := game.Board[sourcerow][sourcecol].Piece

	//clear the source square
	game.Board[sourcerow][sourcecol] = Square{
		Occupied: false,
		Piece:    nil,
	}

	squareOccupied, _ := Occupied_squares(*game, move_pos)
	if squareOccupied {
		fmt.Printf("square occupied")
		return
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

}
