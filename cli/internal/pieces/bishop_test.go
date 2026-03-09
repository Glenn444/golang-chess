package pieces

import (
	"testing"

	"github.com/Glenn444/golang-chess/utils"
	"github.com/stretchr/testify/require"
)

func TestBishop(t *testing.T) {
	board := [][]Square{}
	gameState := GameState{
		CurrentPlayer: "w",
		Board: Initialise_board(board),
	}
	bishopTests := []struct {
		name         string
		piece        PieceInterface
		legalSquares []string
	}{
		{
			name: "bishop1",
			piece: &Bishop{
				PieceType: "B",
				Color:     "w",
				Position:  "c1",
			},
			legalSquares: []string{"b2", "a3", "d2", "e3", "f4", "g5", "h6"},
		},
		{
			name: "bishop2",
			piece: &Bishop{
				PieceType: "B",
				Color:     "w",
				Position:  "d4",
			},
			legalSquares: []string{"c3", "b2", "a1", "e5", "f6", "g7", "h8", "c5", "b6", "a7", "e3", "f2", "g1"},
		},
		{
			name: "bishop3",
			piece: &Bishop{
				PieceType: "B",
				Color:     "b",
				Position:  "a1",
			},
			legalSquares: []string{"b2", "c3", "d4", "e5", "f6", "g7", "h8"},
		},
		{
			name: "bishop4",
			piece: &Bishop{
				PieceType: "B",
				Color:     "b",
				Position:  "h8",
			},
			legalSquares: []string{"g7", "f6", "e5", "d4", "c3", "b2", "a1"},
		},
	}


	for _, tt := range bishopTests {
		t.Run(tt.name, func(t *testing.T) {
			gotlegalSquares := tt.piece.GetLegalSquares()

			require.ElementsMatch(t,gotlegalSquares,tt.legalSquares)
			
		})
	}
}



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
			case "P", "p":
				if b[pos] == "P" {
					board[i][j] = Square{
						Occupied: true,
						Piece: &Pawn{
							Color:     "w",
							PieceType: "P",
							Position:  pos,
						},
					}
				} else {
					board[i][j] = Square{
						Occupied: true,
						Piece: &Pawn{
							Color:     "b",
							PieceType: "P",
							Position:  pos,
						},
					}
				}
			case "R", "r":
				if b[pos] == "R" {
					board[i][j] = Square{
						Occupied: true,
						Piece: &Rook{
							Color:     "w",
							PieceType: "R",
							Position:  pos,
						},
					}
				} else {
					board[i][j] = Square{
						Occupied: true,
						Piece: &Rook{
							Color:     "b",
							PieceType: "R",
							Position:  pos,
						},
					}
				}
			case "N", "n":
				if b[pos] == "N" {
					board[i][j] = Square{
						Occupied: true,
						Piece: &Knight{
							Color:     "w",
							PieceType: "N",
							Position:  pos,
						},
					}
				} else {
					board[i][j] = Square{
						Occupied: true,
						Piece: &Knight{
							Color:     "b",
							PieceType: "N",
							Position:  pos,
						},
					}
				}

			case "B", "b":
				if b[pos] == "B" {
					board[i][j] = Square{
						Occupied: true,
						Piece: &Bishop{
							Color:     "w",
							PieceType: "B",
							Position:  pos,
						},
					}
				} else {
					board[i][j] = Square{
						Occupied: true,
						Piece: &Bishop{
							Color:     "b",
							PieceType: "B",
							Position:  pos,
						},
					}
				}
			case "Q", "q":
				if b[pos] == "Q" {
					board[i][j] = Square{
						Occupied: true,
						Piece: &pieces.Queen{
							Color:     "w",
							PieceType: "Q",
							Position:  pos,
						},
					}
				} else {
					board[i][j] = Square{
						Occupied: true,
						Piece: &pieces.Queen{
							Color:     "b",
							PieceType: "Q",
							Position:  pos,
						},
					}
				}
			case "K", "k":
				if b[pos] == "K" {
					board[i][j] = Square{
						Occupied: true,
						Piece: &pieces.King{
							Color:     "w",
							PieceType: "K",
							Position:  pos,
						},
					}
				} else {
					board[i][j] = Square{
						Occupied: true,
						Piece: &pieces.King{
							Color:     "b",
							PieceType: "K",
							Position:  pos,
						},
					}
				}

			}
		}
	}
	return board

}