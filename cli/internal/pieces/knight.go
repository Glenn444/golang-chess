package pieces

import (
	"github.com/Glenn444/golang-chess/utils"
)

type Knight struct {
	PieceType string
	Color     string
	Position  string
	Points int64
}

// current_position = "a2"
func (k *Knight) GetLegalSquares(g GameState) []string {

	var possible_positions []string
	//letter := string(k.Position[0])
	//num, _ := strconv.Atoi(k.Position[1:])

	row, col := utils.Chess_notation_to_indices(k.Position)

	knightMoves := [][2]int{{2, 1}, {2, -1}, {-2, 1}, {-2, -1}, {1, 2}, {1, -2}, {-1, 2}, {-1, -2}}

	for _, k_move := range knightMoves {
		new_row := row + k_move[0]
		new_col := col + k_move[1]

		if new_row >= 0 && new_row < 8 && new_col >= 0 && new_col < 8 {
			pos := utils.Indices_to_chess_notation(new_row, new_col)
			if g.Board[new_row][new_col].Occupied && g.Board[new_row][new_col].Piece.GetColor() != k.Color {

				possible_positions = append(possible_positions, pos)
			} else if !g.Board[new_row][new_col].Occupied {
				possible_positions = append(possible_positions, pos)
			}
		}
	}

	return possible_positions
}

func (k *Knight) GetColor() string {
	return k.Color
}

func (k *Knight) GetPosition() string {
	return k.Position
}

func (k *Knight) GetPieceType() string {
	return k.PieceType
}

func (k *Knight) AssignPosition(pos string) {
	k.Position = pos
}
func (k *Knight) GetPiecePoints()int64{
	return k.Points
}

func (k *Knight)Clone()PieceInterface{
	 
	return &King{
		Color: k.Color,
		PieceType: k.PieceType,
		Position: k.Position,
		Points: k.Points,
	}
}
func (k *Knight) String() string {
	if k.Color == "w" {
		return "[♘]" // or "wN"
	}
	return "[♞]" // or "bN"
}
