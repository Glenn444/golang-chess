package board

import (
	"errors"
	"fmt"

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
		switch move {
		case "O-O", "e1g1":
			fmt.Printf("O-O\n")
			//ensure rules of castling hold

			f1 := !gameState.Board[0][5].Occupied
			g1 := !gameState.Board[0][6].Occupied
			rookSquareOccupied := gameState.Board[0][7].Occupied && gameState.Board[0][7].Piece.GetPieceType() == "R"
			kingSquareOccupied := gameState.Board[0][4].Occupied && gameState.Board[0][4].Piece.GetPieceType() == "K"
			if f1 && g1 && rookSquareOccupied && kingSquareOccupied && !gameState.Castle.WhiteKingMoved &&
				!gameState.Castle.WhiteRookKingsideMoved {
				kingPos := utils.Indices_to_chess_notation(0, 6)
				rookPos := utils.Indices_to_chess_notation(0, 5)
				gameState.Board[0][6] = pieces.Square{
					Occupied: true,
					Piece: &pieces.King{
						PieceType: "K",
						Color:     gameState.CurrentPlayer,
						Position:  kingPos,
					},
				}

				gameState.Board[0][5] = pieces.Square{
					Occupied: true,
					Piece: &pieces.Rook{
						PieceType: "R",
						Color:     gameState.CurrentPlayer,
						Position:  rookPos,
					},
				}

				//clear initial King Position and Rook Position
				gameState.Board[0][4] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}

				gameState.Board[0][7] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}
				castled = true
				gameState.StockfishGame = append(gameState.StockfishGame, "e1g1")

			}

			//Queenside castling
		case "O-O-O", "e1c1":

			b1 := !gameState.Board[0][1].Occupied
			c1 := !gameState.Board[0][2].Occupied
			d1 := !gameState.Board[0][3].Occupied
			rookSquareOccupied := gameState.Board[0][0].Occupied && gameState.Board[0][0].Piece.GetPieceType() == "R"
			kingSquareOccupied := gameState.Board[0][4].Occupied && gameState.Board[0][4].Piece.GetPieceType() == "K"
			if b1 && c1 && d1 && rookSquareOccupied && kingSquareOccupied && !gameState.Castle.WhiteKingMoved &&
				!gameState.Castle.WhiteRookQueensideMoved {
				kingPos := utils.Indices_to_chess_notation(0, 2)
				rookPos := utils.Indices_to_chess_notation(0, 1)

				gameState.Board[0][2] = pieces.Square{
					Occupied: true,
					Piece: &pieces.King{
						PieceType: "K",
						Color:     gameState.CurrentPlayer,
						Position:  kingPos,
					},
				}

				gameState.Board[0][1] = pieces.Square{
					Occupied: true,
					Piece: &pieces.Rook{
						PieceType: "R",
						Color:     gameState.CurrentPlayer,
						Position:  rookPos,
					},
				}

				//clear initial King Position and Rook Position
				gameState.Board[0][4] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}

				gameState.Board[0][0] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}
				castled = true
				gameState.StockfishGame = append(gameState.StockfishGame, "e1c1")
			}

		}

	case "b":
		//kingside castling
		switch move {
		case "O-O", "e8g8":
			//ensure rules of castling hold

			f8 := !gameState.Board[7][5].Occupied
			g8 := !gameState.Board[7][6].Occupied
			rookSquareOccupied := gameState.Board[7][7].Occupied && gameState.Board[7][7].Piece.GetPieceType() == "R"
			kingSquareOccupied := gameState.Board[7][4].Occupied && gameState.Board[7][4].Piece.GetPieceType() == "K"
			if f8 && g8 && rookSquareOccupied && kingSquareOccupied && !gameState.Castle.BlackKingMoved &&
				!gameState.Castle.BlackRookKingsideMoved {
				kingPos := utils.Indices_to_chess_notation(7, 6)
				rookPos := utils.Indices_to_chess_notation(7, 5)
				gameState.Board[7][6] = pieces.Square{
					Occupied: true,
					Piece: &pieces.King{
						PieceType: "K",
						Color:     gameState.CurrentPlayer,
						Position:  kingPos,
					},
				}

				gameState.Board[7][5] = pieces.Square{
					Occupied: true,
					Piece: &pieces.Rook{
						PieceType: "R",
						Color:     gameState.CurrentPlayer,
						Position:  rookPos,
					},
				}

				//clear initial King Position and Rook Position
				gameState.Board[7][4] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}

				gameState.Board[7][7] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}
				castled = true
				gameState.StockfishGame = append(gameState.StockfishGame, "e8g8")

			}

			//Queenside castling
		case "O-O-O", "e8c8":
			
			b8 := !gameState.Board[7][1].Occupied
			c8 := !gameState.Board[7][2].Occupied
			d8 := !gameState.Board[7][3].Occupied
			rookSquareOccupied := gameState.Board[7][0].Occupied && gameState.Board[7][0].Piece.GetPieceType() == "R"
			kingSquareOccupied := gameState.Board[7][4].Occupied && gameState.Board[7][4].Piece.GetPieceType() == "K"
			if b8 && c8 && d8 && rookSquareOccupied && kingSquareOccupied && !gameState.Castle.BlackKingMoved &&
				!gameState.Castle.BlackRookQueensideMoved {
				kingPos := utils.Indices_to_chess_notation(7, 2)
				rookPos := utils.Indices_to_chess_notation(7, 1)
				gameState.Board[7][2] = pieces.Square{
					Occupied: true,
					Piece: &pieces.King{
						PieceType: "K",
						Color:     gameState.CurrentPlayer,
						Position:  kingPos,
					},
				}

				gameState.Board[7][1] = pieces.Square{
					Occupied: true,
					Piece: &pieces.Rook{
						PieceType: "R",
						Color:     gameState.CurrentPlayer,
						Position:  rookPos,
					},
				}

				//clear initial King Position and Rook Position
				gameState.Board[7][4] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}

				gameState.Board[7][0] = pieces.Square{
					Occupied: false,
					Piece:    nil,
				}
				castled = true
				gameState.StockfishGame = append(gameState.StockfishGame, "e8c8")
			}
		}
	}
	if !castled {
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
