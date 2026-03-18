package board

import (
	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/utils"
)

func Move(game *pieces.GameState, move string) error {
	var move_pos string
	move_pos = string(move[1:])
	moveType := string(move[1])


	//the game is in check
	if game.Check{

	}
	if moveType == "x" || moveType == "X"{
		err := CapturePiece(game,move)
		if err != nil{
			return err
		}
		return nil

	}else if len(move) == 4{
		move_pos = string(move[2:])
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

	//checking check
	pieceLegalSquares := piece.GetLegalSquares(*game)
	for _,squares := range game.Board{
		for _,square := range squares{
			if square.Occupied && square.Piece.GetPieceType() == "K" && square.Piece.GetColor() != game.CurrentPlayer{
				for _,legalPos := range pieceLegalSquares{
					if square.Piece.GetPosition() == legalPos{
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

	return nil
}
