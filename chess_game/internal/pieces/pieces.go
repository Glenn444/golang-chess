package pieces

import (
	"fmt"
	"strings"

	"github.com/Glenn444/golang-chess/internal/utils/chess"
)

type Square struct {
	Occupied bool
	Piece    PieceInterface
}

type Castling struct {
	WhiteKingMoved          bool
	WhiteRookKingsideMoved  bool
	WhiteRookQueensideMoved bool
	BlackKingMoved          bool
	BlackRookKingsideMoved  bool
	BlackRookQueensideMoved bool
}

type GameState struct {
	CurrentPlayer  string
	Board          [][]Square
	CapturedPieces map[string][]PieceInterface
	StockfishGame  []string
	PlayAgainst    string //person or stockfish
	UserColor      string
	Castle	Castling
}

type PieceInterface interface {
	GetLegalSquares(g GameState) []string
	GetColor() string
	GetPosition() string
	GetPieceType() string
	AssignPosition(pos string)
	String() string
	GetPiecePoints() int64
	Clone() PieceInterface
}

func PrintBoard(initialBoard_position [][]Square) {

	fmt.Printf("      a  b  c  d  e  f  g  h\n")
	for i, row := range initialBoard_position {

		fmt.Printf("%d", i+1)
		fmt.Printf("    ")
		for _, s := range row {

			if s.Occupied {
				fmt.Printf("%v", s.Piece.String())
			} else {
				fmt.Printf("[ ]")
			}

		}
		fmt.Printf("\n")

	}
	fmt.Printf("      a  b  c  d  e  f  g  h\n")
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

func SetUpBoard(b map[string]string) [][]Square {
	chessBoard := Create_board()

	for i, row := range chessBoard {
		for j := range row {
			pos := utils.Indices_to_chess_notation(i, j)
			color := "b"
			if b[pos] == strings.ToUpper(b[pos]) {
				color = "w"
			}
			switch b[pos] {
			case "P", "p":
				chessBoard[i][j] = Square{
					Occupied: true,
					Piece: &Pawn{
						Color:     color,
						PieceType: strings.ToUpper(b[pos]),
						Position:  pos,
						Points:    1,
					},
				}
			case "R", "r":
				chessBoard[i][j] = Square{
					Occupied: true,
					Piece: &Rook{
						Color:     color,
						PieceType: strings.ToUpper(b[pos]),
						Position:  pos,
						Points:    5,
					},
				}
			case "N", "n":
				chessBoard[i][j] = Square{
					Occupied: true,
					Piece: &Knight{
						Color:     color,
						PieceType: strings.ToUpper(b[pos]),
						Position:  pos,
						Points:    3,
					},
				}

			case "B", "b":
				chessBoard[i][j] = Square{
					Occupied: true,
					Piece: &Bishop{
						Color:     color,
						PieceType: strings.ToUpper(b[pos]),
						Position:  pos,
						Points:    3,
					},
				}
			case "Q", "q":
				chessBoard[i][j] = Square{
					Occupied: true,
					Piece: &Queen{
						Color:     color,
						PieceType: strings.ToUpper(b[pos]),
						Position:  pos,
						Points:    9,
					},
				}
			case "K", "k":

				chessBoard[i][j] = Square{
					Occupied: true,
					Piece: &King{
						Color:     color,
						PieceType: strings.ToUpper(b[pos]),
						Position:  pos,
					},
				}

			}
		}
	}
	return chessBoard

}
