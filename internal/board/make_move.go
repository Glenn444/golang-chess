package board

import (
	"fmt"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/utils"
)

func Move(game *GameState, move string) {
	//pieces := []string{"Q","N","K","R","B"}
	pieceType := string(move[0])
	move_pos := string(move[1:])
	row, col := utils.Chess_notation_to_indices(move_pos)
	pos := CurrentPlayer_Occupied_Piece_position(*game, move)
	//row, col := utils.Chess_notation_to_indices(pos)

	fmt.Printf("Piece Position: %s\n", pos)

	squareOccupied,_ := Occupied_squares(*game,move_pos)
	if squareOccupied{
		fmt.Printf("square occupied")
		return
	}
	switch pieceType {
	case "Q":
		game.Board[row][col] = Square{
			Occupied: false,
			Piece: pieces.Queen{
				PieceType: "Q",
				Color:     game.CurrentPlayer,
				Position:  move_pos,
			},
		}
	case "N":
		game.Board[row][col] = Square{
			Occupied: false,
			Piece: pieces.Knight{
				PieceType: "N",
				Color:     game.CurrentPlayer,
				Position:  move_pos,
			},
		}
	case "K":
		game.Board[row][col] = Square{
			Occupied: false,
			Piece: pieces.King{
				PieceType: "K",
				Color:     game.CurrentPlayer,
				Position:  move_pos,
			}}
	case "R":
		game.Board[row][col] = Square{
			Occupied: false,
			Piece: pieces.Rook{
				PieceType: "R",
				Color:     game.CurrentPlayer,
				Position:  move_pos,
			}}
	case "B":
		game.Board[row][col] = Square{
			Occupied: false,
			Piece: pieces.Bishop{
				PieceType: "B",
				Color:     game.CurrentPlayer,
				Position:  move_pos,
			}}
	default:

	}

}

// func Get_Piece()  {

// }
