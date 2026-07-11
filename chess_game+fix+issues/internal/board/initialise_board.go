package board

import (
	"strings"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/internal/utils/chess"
)

func Initialise_board(board [][]pieces.Square) [][]pieces.Square {
	b := map[string]string{
		// White pieces
		"a2": "P", "b2": "P", "c2": "P", "d2": "P", "e2": "P", "f2": "P", "g2": "P", "h2": "P",
		"a1": "R", "h1": "R",
		"b1": "N", "g1": "N",
		"c1": "B", "f1": "B",
		"d1": "Q",
		"e1": "K",

		// Black pieces
		"a7": "p", "b7": "p", "c7": "p", "d7": "p", "e7": "p", "f7": "p", "g7": "p", "h7": "p",
		"a8": "r", "h8": "r",
		"b8": "n", "g8": "n",
		"c8": "b", "f8": "b",
		"d8": "q",
		"e8": "k",
	}
	

	for i, row := range board {
		for j := range row {
			pos := chess.Indices_to_chess_notation(i, j)
			color := "b"
			if b[pos] == strings.ToUpper(b[pos]) {
				color = "w"
			}
			switch b[pos] {
			case "P", "p":
				board[i][j] = pieces.Square{
					Occupied: true,
					Piece: &pieces.Pawn{
						Color:     color,
						PieceType: strings.ToUpper(b[pos]),
						Position:  pos,
						Points: 1,
					},
				}
			case "R", "r":
				board[i][j] = pieces.Square{
					Occupied: true,
					Piece: &pieces.Rook{
						Color:     color,
						PieceType: strings.ToUpper(b[pos]),
						Position:  pos,
						Points: 5,
					},
				}
			case "N", "n":
				board[i][j] = pieces.Square{
					Occupied: true,
					Piece: &pieces.Knight{
						Color:     color,
						PieceType: strings.ToUpper(b[pos]),
						Position:  pos,
						Points: 3,
					},
				}

			case "B", "b":
				board[i][j] = pieces.Square{
					Occupied: true,
					Piece: &pieces.Bishop{
						Color:     color,
						PieceType: strings.ToUpper(b[pos]),
						Position:  pos,
						Points: 3,
					},
				}
			case "Q", "q":
				board[i][j] = pieces.Square{
					Occupied: true,
					Piece: &pieces.Queen{
						Color:     color,
						PieceType: strings.ToUpper(b[pos]),
						Position:  pos,
						Points: 9,
					},
				}
			case "K", "k":

				board[i][j] = pieces.Square{
					Occupied: true,
					Piece: &pieces.King{
						Color:     color,
						PieceType: strings.ToUpper(b[pos]),
						Position:  pos,
					},
				}

			}
		}
	}
	return board

}
