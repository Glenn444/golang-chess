package board

import (
	"errors"
	"fmt"

	"strconv"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/internal/utils/chess"
)

func CapturePiece(game *pieces.GameState, move string) error {
	//var initialPiecePosition string
	//var coordinateMove string
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

			//initialPiecePosition = initialPos

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
			stockfishMove := fmt.Sprintf("%s%s", initialPos, destCapturePos)
			game.StockfishGame = append(game.StockfishGame, stockfishMove)
			fmt.Printf("move added to stockfish game: %s %s",game.StockfishGame, stockfishMove)

			//stockfish coordinate move
			//coordinateMove := fmt.Sprintf("%s%s",initialPos,destPiece.GetPosition())
			return nil

		} else if pieceType == "N" || pieceType == "Q" || pieceType == "K" || pieceType == "B" || pieceType == "R" {
			initialPiece, err := GetInitialPositionByPiece(destCapturePos, pieceType, *game)
			if err != nil {
				return errors.New("invalid move capture")
			}

			initialPos := initialPiece.GetPosition()

			//initialPiecePosition = initialPos
			//type of move Ra1,Rh8 or Ke1 or Ke8 to check for castling rules
			movedPieceType := fmt.Sprintf("%s%s", initialPiece.GetPieceType(), initialPos)
			pieces.CastlePieceMoved(game, movedPieceType)

			
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
			
			stockfishMove := fmt.Sprintf("%s%s", initialPos, destCapturePos)
			game.StockfishGame = append(game.StockfishGame, stockfishMove)
			fmt.Printf("move added to stockfish game: %s %s",game.StockfishGame, stockfishMove)

			//coordinateMove := fmt.Sprintf("%s%s",initialPos,destPiece.GetPosition())
			return nil
		}
	}
	return errors.New("capture piece error occurred")
}
