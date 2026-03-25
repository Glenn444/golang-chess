package board

import (
	"errors"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/utils"
)

// O-O Kingside castling
// O-O-O Queenside castling
func CastlingMove(gameState *pieces.GameState, move string) error {
	var castled bool
	switch gameState.CurrentPlayer {
	case "w":
		//kingside castling
		if move == "O-O" || move == "e1g1" {
			//ensure rules of castling hold

			f1 := !gameState.Board[1][6].Occupied
			g1 := !gameState.Board[1][7].Occupied
			rookSquareOccupied := gameState.Board[1][8].Occupied && gameState.Board[1][8].Piece.GetPieceType() == "R"
			kingSquareOccupied := gameState.Board[1][5].Occupied && gameState.Board[1][5].Piece.GetPieceType() == "K"
			if f1 && g1 && rookSquareOccupied && kingSquareOccupied {
				kingPos := utils.Indices_to_chess_notation(1, 7)
				rookPos := utils.Indices_to_chess_notation(1, 6)
				gameState.Board[1][7] = pieces.Square{
					Occupied: true,
					Piece: &pieces.King{
						PieceType: "K",
						Color:     gameState.CurrentPlayer,
						Position:  kingPos,
					},
				}

				gameState.Board[1][6] = pieces.Square{
					Occupied: true,
					Piece: &pieces.Rook{
						PieceType: "R",
						Color:     gameState.CurrentPlayer,
						Position:  rookPos,
					},
				}

				//clear initial King Position and Rook Position
				gameState.Board[1][5] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}

				gameState.Board[1][8] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}
				castled = true

			}
			//Queenside castling
		} else if move == "O-O-O" || move == "e1c1" {

			b1 := !gameState.Board[1][2].Occupied
			c1 := !gameState.Board[1][3].Occupied
			d1 := !gameState.Board[1][4].Occupied
			rookSquareOccupied := gameState.Board[1][1].Occupied && gameState.Board[1][1].Piece.GetPieceType() == "R"
			kingSquareOccupied := gameState.Board[1][5].Occupied && gameState.Board[1][5].Piece.GetPieceType() == "K"
			if b1 && c1 && d1 && rookSquareOccupied && kingSquareOccupied {
				kingPos := utils.Indices_to_chess_notation(1, 3)
				rookPos := utils.Indices_to_chess_notation(1, 2)

				gameState.Board[1][3] = pieces.Square{
					Occupied: true,
					Piece: &pieces.King{
						PieceType: "K",
						Color:     gameState.CurrentPlayer,
						Position:  kingPos,
					},
				}

				gameState.Board[1][2] = pieces.Square{
					Occupied: true,
					Piece: &pieces.Rook{
						PieceType: "R",
						Color:     gameState.CurrentPlayer,
						Position:  rookPos,
					},
				}

				//clear initial King Position and Rook Position
				gameState.Board[1][5] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}

				gameState.Board[1][1] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}
				castled = true
			}

		}

	case "b":
		//kingside castling
		switch move {
		case "O-O", "e8g8":
			//ensure rules of castling hold

			f8 := !gameState.Board[8][6].Occupied
			g8 := !gameState.Board[8][7].Occupied
			rookSquareOccupied := gameState.Board[8][8].Occupied && gameState.Board[8][8].Piece.GetPieceType() == "R"
			kingSquareOccupied := gameState.Board[8][5].Occupied && gameState.Board[8][5].Piece.GetPieceType() == "K"
			if f8 && g8 && rookSquareOccupied && kingSquareOccupied {
				kingPos := utils.Indices_to_chess_notation(8, 7)
				rookPos := utils.Indices_to_chess_notation(8, 6)
				gameState.Board[8][7] = pieces.Square{
					Occupied: true,
					Piece: &pieces.King{
						PieceType: "K",
						Color:     gameState.CurrentPlayer,
						Position:  kingPos,
					},
				}

				gameState.Board[8][6] = pieces.Square{
					Occupied: true,
					Piece: &pieces.Rook{
						PieceType: "R",
						Color:     gameState.CurrentPlayer,
						Position:  rookPos,
					},
				}

				//clear initial King Position and Rook Position
				gameState.Board[8][5] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}

				gameState.Board[8][8] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}
				castled = true

			}
			//Queenside castling
		case "O-O-O", "e8c8":

			b8 := !gameState.Board[8][2].Occupied
			c8 := !gameState.Board[8][3].Occupied
			d8 := !gameState.Board[8][4].Occupied
			rookSquareOccupied := gameState.Board[8][1].Occupied && gameState.Board[8][1].Piece.GetPieceType() == "R"
			kingSquareOccupied := gameState.Board[8][5].Occupied && gameState.Board[8][5].Piece.GetPieceType() == "K"
			if b8 && c8 && d8 && rookSquareOccupied && kingSquareOccupied {
				kingPos := utils.Indices_to_chess_notation(8, 3)
				rookPos := utils.Indices_to_chess_notation(8, 2)
				gameState.Board[8][3] = pieces.Square{
					Occupied: true,
					Piece: &pieces.King{
						PieceType: "K",
						Color:     gameState.CurrentPlayer,
						Position:  kingPos,
					},
				}

				gameState.Board[8][2] = pieces.Square{
					Occupied: true,
					Piece: &pieces.Rook{
						PieceType: "R",
						Color:     gameState.CurrentPlayer,
						Position:  rookPos,
					},
				}

				//clear initial King Position and Rook Position
				gameState.Board[8][5] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}

				gameState.Board[8][1] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}
				castled = true
			}
		}
	}
	if !castled{
		return errors.New("invalid castling move")
	}
	//change current player after making move
	if gameState.CurrentPlayer == "w" {
		gameState.CurrentPlayer = "b"
	} else {
		gameState.CurrentPlayer = "w"
	}

	
return nil
}
