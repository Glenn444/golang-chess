package pieces

import "github.com/Glenn444/golang-chess/utils"

type Square struct {
	Occupied bool
	Piece    PieceInterface
}

type GameState struct {
	CurrentPlayer string
	Board         [][]Square
}

type PieceInterface interface{
	GetLegalSquares(g GameState) []string
	GetColor() string
    GetPosition() string
    GetPieceType() string
	AssignPosition(pos string)
	String() string
}

func Create_board() [][]Square {

	rows, cols := 8, 8

	board := make([][]Square, rows)

	for i := range board {
		board[i] = make([]Square, cols)
	}

	// First, initialize all squares as empty
	for i := range board {
		for j := range board[i] {
			board[i][j] = Square{
				Occupied: false,
				Piece:    nil,
			}
		}
	}

	return board
}
func Initialise_board(board [][]Square) [][]Square {
	b := map[string]string{
		// White pieces
		"a2": "P", "b3": "P", "c2": "P", "d2": "P", "e2": "P", "f2": "P", "g2": "P", "h2": "P",
		"a1": "R", "h1": "R",
		"b1": "N", "g1": "N",
		"b2": "B", "f1": "B",
		"d1": "Q",
		"e1": "K",

		// Black pieces
		"a7": "p", "b7": "p", "c7": "p", "d7": "p", "e5": "p", "f7": "p", "g7": "p", "h7": "p",
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
						Piece: &Queen{
							Color:     "w",
							PieceType: "Q",
							Position:  pos,
						},
					}
				} else {
					board[i][j] = Square{
						Occupied: true,
						Piece: &Queen{
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
						Piece: &King{
							Color:     "w",
							PieceType: "K",
							Position:  pos,
						},
					}
				} else {
					board[i][j] = Square{
						Occupied: true,
						Piece: &King{
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
