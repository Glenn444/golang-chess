package board

import (
	"strings"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/utils"
)

func Initialise_board(board [][]Square) [][]Square {
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
			pos := utils.Indices_to_chess_notation(i, j)
			switch b[pos] {
			case "P", "q":
				if b[pos] == "P" {
					board[i][j] = Square{
						Occupied: true,
						Piece: pieces.Pawn{
							Color:     "white",
							PieceType: "P",
							Position:  pos,
						},
					}
				}else{
					board[i][j] = Square{
						Occupied: true,
						Piece: piece{
							Color:     "black",
							PieceType: "P",
							Position:  pos,
						},
					}
				}
			case "p", "r", "n", "k", "b":
				board[i][j] = Square{
					Occupied: true,
					Piece: piece{
						Color:     "black",
						PieceType: strings.ToUpper(b[pos]),
						Position:  pos,
					},
				}
			default:
				board[i][j] = Square{
					Occupied: false,
					Piece: piece{
						Color:     "",
						PieceType: "",
						Position:  "",
					},
				}
			}
		}
	}

	return board

}
