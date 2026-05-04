package pieces

import (

	"github.com/Glenn444/golang-chess/internal/utils/chess"
)

type Rook struct {
	PieceType string
	Color     string
	Position  string
	Points	int64
}

func (r *Rook) GetLegalSquares(g GameState) []string {
	var positions []string

	allDiagnols := [][]string{
		getHorizontalVertical(*r, -1, 0),
		getHorizontalVertical(*r, 0, -1),
		getHorizontalVertical(*r, 0, 1),
		getHorizontalVertical(*r, 1, 0),
	}
	for _, diag := range allDiagnols {
		for _, pos := range diag {
			i, j, _ := chess.ChessNotationToIndices(pos)
			square := g.Board[i][j]

			if square.Occupied {
				if square.Piece.GetColor() != r.Color {
					positions = append(positions, pos)
				}
				break
			}

			positions = append(positions, pos)
		}

	}
	return positions

}

func getHorizontalVertical(r Rook, rowDelta int, colDelta int) []string {
	var possible_positions []string

	row, col, _ := chess.ChessNotationToIndices(r.Position)

	for {
		r := row + rowDelta
		c := col + colDelta

		if r >= 8 || r < 0 || c >= 8 || c < 0 {
			break
		}
		pos := chess.Indices_to_chess_notation(r, c)
		possible_positions = append(possible_positions, pos)
		row = r
		col = c
	}
	return possible_positions
}

func (r *Rook) GetColor() string {
	return r.Color
}

func (r *Rook) GetPosition() string {
	return r.Position
}

func (r Rook) GetPieceType() string {
	return r.PieceType
}

func (r *Rook) AssignPosition(pos string) {
	r.Position = pos
}
func (r *Rook) GetPiecePoints()int64{
	return r.Points
}

func (r *Rook)Clone()PieceInterface{
	 
	return &Rook{
		Color: r.Color,
		PieceType: r.PieceType,
		Position: r.Position,
		Points: r.Points,
	}
}
func (r *Rook) String() string {
	if r.Color == "w" {
		return "[♖]" // or "wR"
	}
	return "[♜]" // or "bR"
}
