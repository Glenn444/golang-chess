package board

import (
	"errors"
	"fmt"
	"slices"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/utils"
)

func Move(game1 *pieces.GameState, move string) error {
	var stockfishMove string
	var move_pos string


	if utils.IsAlgebraic(move){
		move1,err := CoordinateToAlgebraic(*game1,move)
		if err !=nil{
			return err
		}
		move = move1
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

	} else if len(move) == 4 && slices.Contains([]string{"B","N","Q","R"},string(move[0])){
		// piece types are: R,B,N,Q
		//when move is Rhe1 or R1e4
		move_pos = string(move[2:])
	}

	sourcepos, err := CurrentPlayer_Occupied_Piece_position(*game, move)
	if err != nil {
		return err
	}

	
	//pawn move
	if len(move) == 2 {
		move_pos = move
	}

	destrow, destcol := utils.Chess_notation_to_indices(move_pos)
	sourcerow, sourcecol := utils.Chess_notation_to_indices(sourcepos)

	piece := game.Board[sourcerow][sourcecol].Piece
	piece.AssignPosition(move_pos)

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

	stockfishMove = fmt.Sprintf("%s%s",sourcepos,move)
	game.StockfishGame = append(game.StockfishGame, stockfishMove)
	//change current player after making move
	if game1.CurrentPlayer == "w" {
		game1.CurrentPlayer = "b"
	} else {
		game1.CurrentPlayer = "w"
	}

	return nil
}
