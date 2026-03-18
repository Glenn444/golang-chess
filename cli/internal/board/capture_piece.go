package board

import (
	"errors"
	"fmt"

	"strconv"

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
		//black capture and white capture.
		var initialCapturePosNum int
		if game.CurrentPlayer == "w" {
			numPos, _ := strconv.Atoi(string(move[3]))
			initialCapturePosNum = numPos - 1
		} else {
			numPos, _ := strconv.Atoi(string(move[3]))
			initialCapturePosNum = numPos + 1
		}
		if boardFile[pieceType] {
			//this is a pawn capture

			initialPos := fmt.Sprintf("%s%d", pieceType, initialCapturePosNum)
			destrow, destcol := utils.Chess_notation_to_indices(destCapturePos)
			sourcerow, sourcecol := utils.Chess_notation_to_indices(initialPos)

			piece := game.Board[sourcerow][sourcecol].Piece
			piece.AssignPosition(destCapturePos)

			//clear the source square
			game.Board[sourcerow][sourcecol] = pieces.Square{
				Occupied: false,
				Piece:    nil,
			}

			//add captured pieces to current player
			destPiece := game.Board[destrow][destcol].Piece

			game.CapturedPieces[game.CurrentPlayer] = append(game.CapturedPieces[game.CurrentPlayer], destPiece)
			//destination square
			game.Board[destrow][destcol] = pieces.Square{
				Occupied: true,
				Piece:    piece,
			}

			//change current player after making move
			if game.CurrentPlayer == "w" {
				game.CurrentPlayer = "b"
			} else {
				game.CurrentPlayer = "w"
			}

		} else if pieceType == "N" || pieceType == "Q" || pieceType == "K" || pieceType == "B" || pieceType == "R" {
			initialPiece, err := GetInitialPositionByPiece(destCapturePos, pieceType, *game)
			if err != nil {
				return errors.New("invalid move capture")
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
			destPiece := game.Board[destrow][destcol].Piece

			game.CapturedPieces[game.CurrentPlayer] = append(game.CapturedPieces[game.CurrentPlayer], destPiece)
			//destination square
			game.Board[destrow][destcol] = pieces.Square{
				Occupied: true,
				Piece:    piece,
			}
			pieceLegalSquares := piece.GetLegalSquares(*game)
			for _, squares := range game.Board {
				for _, square := range squares {
					if square.Occupied && square.Piece.GetPieceType() == "K" && square.Piece.GetColor() != game.CurrentPlayer {
						for _, legalPos := range pieceLegalSquares {
							if square.Piece.GetPosition() == legalPos {
								game.Check = true
							}
						}
					}
				}
			}

			//change current player after making move
			if game.CurrentPlayer == "w" {
				game.CurrentPlayer = "b"
			} else {
				game.CurrentPlayer = "w"
			}

		}
	}
	return nil
}
