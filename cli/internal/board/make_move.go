package board

import (
	"errors"
	"fmt"
	"slices"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/utils"
)

func Move(game1 *pieces.GameState, move string) error {
	var appended bool
	var move_pos string

	if utils.IsaCastlingMove(move) {
		fmt.Printf("its a castling move\n")
		castlingErr := CastlingMove(game1, move)
		if castlingErr != nil {
			return castlingErr
		}
		return nil
	}
	if utils.IsAlgebraic(move) {
		algebraicMove, err := CoordinateToAlgebraic(*game1, move)
		if err != nil {
			return err
		}
		//append the move directly if it's coordinate
		game1.StockfishGame = append(game1.StockfishGame, move)
		appended = true

		move = algebraicMove

	}
	move_pos = string(move[1:])
	moveType := string(move[1])

	boardA := Create_board()
	CopyBoard(boardA, game1.Board)
	game := &pieces.GameState{
		CurrentPlayer:  game1.CurrentPlayer,
		Board:          boardA,
		CapturedPieces: make(map[string][]pieces.PieceInterface),
	}

	if moveType == "x" || moveType == "X" {
		err := CapturePiece(game, move)
		if err != nil {
			return err
		}
		if IsKinginCheck(*game) {
			return errors.New("King is still in check!!!\n")
		} else {

			CopyBoard(game1.Board, game.Board)
			capturedSlice := game.CapturedPieces[game.CurrentPlayer]

			game1.CapturedPieces[game.CurrentPlayer] = append(game1.CapturedPieces[game.CurrentPlayer], capturedSlice...)

		}

		//change current player after making move
		if game1.CurrentPlayer == "w" {
			game1.CurrentPlayer = "b"
		} else {
			game1.CurrentPlayer = "w"
		}

		return nil

	} else if len(move) == 4 && slices.Contains([]string{"B", "N", "Q", "R", "K"}, string(move[0])) {
		// piece types are: R,B,N,Q,K
		//when move is Rhe1 or R1e4
		move_pos = string(move[2:])
	}

	sourcepos, err := CurrentPlayer_Occupied_Piece_position(*game, move)
	if err != nil {
		return err
	}

	//initialPiecePosition = sourcepos
	//pawn move
	if len(move) == 2 {
		move_pos = move
	}

	destrow, destcol := utils.Chess_notation_to_indices(move_pos)
	sourcerow, sourcecol := utils.Chess_notation_to_indices(sourcepos)

	piece := game.Board[sourcerow][sourcecol].Piece
	piece.AssignPosition(move_pos)

	//type of move Ra1,Rh8 or Ke1 or Ke8 to check for castling rules
	movedPieceType := fmt.Sprintf("%s%s", piece.GetPieceType(), sourcepos)
	pieces.CastlePieceMoved(game1, movedPieceType)

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

	//checking check
	if IsKinginCheck(*game) {
		return errors.New("king is still in check!!!\n")
	} else {
		CopyBoard(game1.Board, game.Board)
		capturedSlice := game.CapturedPieces[game.CurrentPlayer]

		game1.CapturedPieces[game.CurrentPlayer] = append(game1.CapturedPieces[game.CurrentPlayer], capturedSlice...)
	}

	coordinatePos := fmt.Sprintf("%s%s", sourcepos, move_pos)
	if !appended {
		game1.StockfishGame = append(game1.StockfishGame, coordinatePos)
		fmt.Printf("coordinate Pos: %s\n", coordinatePos)
		fmt.Printf("after playing stockfishGame: %s\n", game1.StockfishGame)
	}

	//change current player after making move
	if game1.CurrentPlayer == "w" {
		game1.CurrentPlayer = "b"
	} else {
		game1.CurrentPlayer = "w"
	}

	return nil
}
