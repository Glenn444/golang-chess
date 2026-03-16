package board

import (
	"errors"
	"fmt"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/utils"
)

func CapturePiece(game *pieces.GameState, move string) error {
	destCapturePos := string(move[2:])
	pieceType := string(move[0])

	boardFile := map[string]bool{
		"a": true, "b": true, "c": true, "d": true, "e": true, "f": true, "g": true, "h": true,
	}

	row, col := utils.Chess_notation_to_indices(destCapturePos)
	pieceSquare := game.Board[row][col]

	if pieceSquare.Occupied && game.CurrentPlayer != pieceSquare.Piece.GetColor() {
		//valid capture
		if boardFile[pieceType] {
			//this is a pawn capture
			initialCapturePosNum := int(move[3]) - 1
			initialPos := fmt.Sprintf("%s%d", pieceType, initialCapturePosNum)
			destrow, destcol := utils.Chess_notation_to_indices(destCapturePos)
			sourcerow, sourcecol := utils.Chess_notation_to_indices(initialPos)

			piece := game.Board[sourcerow][sourcecol].Piece
			piece.AssignPosition(destCapturePos)
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

		}else if pieceType == "N" || pieceType == "Q" || pieceType == "K" || pieceType == "B" || pieceType == "R"{
			initialPiece,err := GetInitialPositionByPiece(destCapturePos,pieceType,*game)
			if err != nil{
				return errors.New("invalid move")
			}

			initialPos := initialPiece.GetPosition()
			destrow, destcol := utils.Chess_notation_to_indices(destCapturePos)
			sourcerow, sourcecol := utils.Chess_notation_to_indices(initialPos)

			piece := game.Board[sourcerow][sourcecol].Piece
			piece.AssignPosition(destCapturePos)
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

		}
	}
	return errors.New("error happened and i don't know")
}
