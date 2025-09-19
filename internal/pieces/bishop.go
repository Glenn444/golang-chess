package pieces

import (
	"fmt"

	"github.com/Glenn444/golang-chess/utils"
)

var board_letters = []string{"a", "b", "c", "d", "e", "f", "g", "h"}
var nums = []int{1, 2, 3, 4, 5, 6, 7, 8}

type Bishop struct{
	PieceType string
	Color string
	Position string
}

func (b Bishop) GetLegalSquares() []string {
	var positions []string
	pos_top_left := get_horizontal_squares_top_left(b.Position)
	pos_top_right := get_horizontal_squares_top_right(b.Position)

	positions = append(positions, pos_top_left...)
	positions = append(positions, pos_top_right...)

	return positions
}
func get_horizontal_squares_top_left(pos string) []string {

	row, col := utils.Chess_notation_to_indices(pos)
	diagnol := row - col
	var possible_possitions []string

	for _, v := range board_letters {
		for _, j := range nums {
			position := fmt.Sprintf("%s%d", v, j)
			row, col := utils.Chess_notation_to_indices(position)
			diag := row - col

			if diagnol == diag {
				possible_possitions = append(possible_possitions, position)
			}
		}
	}

	return possible_possitions
}

func get_horizontal_squares_top_right(pos string) []string {

	row, col := utils.Chess_notation_to_indices(pos)
	diagnol := row + col
	var possible_possitions []string

	for _, v := range board_letters {
		for _, j := range nums {
			position := fmt.Sprintf("%s%d", v, j)
			row, col := utils.Chess_notation_to_indices(position)
			diag := row + col

			if diagnol == diag {
				possible_possitions = append(possible_possitions, position)
			}
		}
	}

	return possible_possitions
}


func (b Bishop) GetColor() string {
    return b.Color
}

func (b Bishop) GetPosition() string {
    return b.Position
}

func (b Bishop) GetPieceType() string {
    return b.PieceType
}