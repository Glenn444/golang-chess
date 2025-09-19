package pieces

import (
	"fmt"
	"strconv"


	"github.com/Glenn444/golang-chess/utils"
)

type Knight struct {
	PieceType string
	Color     string
	Position  string
}

var boardIndex = map[string]int{
	"a": 0,
	"b": 1,
	"c": 2,
	"d": 3,
	"e": 4,
	"f": 5,
	"g": 6,
	"h": 7,
}
var boardLetters = []string{"a", "b", "c", "d", "e", "f", "g", "h"}


// current_position = "a2"
func (k Knight)GetLegalSquares() []string {

	var possible_positions []string
	letter := string(k.Position[0])
	num, _ := strconv.Atoi(k.Position[1:])

	row, col := utils.Chess_notation_to_indices(k.Position)

	row_top := row + 2
	row_bottom := row - 2
	col_left := col - 2
	col_right := col + 2

	top_pos := fmt.Sprintf("%s%d", letter, row_top)
	bottom_pos := fmt.Sprintf("%s%d", letter, row_bottom)

	top1, top2 := get_squares_along_row(top_pos)
	bottom1, bottom2 := get_squares_along_row(bottom_pos)

	col_left_letter := boardLetters[col_left]
	col_right_letter := boardLetters[col_right]

	left_pos := fmt.Sprintf("%s%d", col_left_letter, num)
	right_pos := fmt.Sprintf("%s%d", col_right_letter, num)

	left1, left2 := get_squares_along_column(left_pos)
	right1, right2 := get_squares_along_column(right_pos)

	possible_positions = append(possible_positions, top1, top2, bottom1, bottom2, left1, left2, right1, right2)

	return possible_positions
}

func get_squares_along_column(position string) (string, string) {
	letter := string(position[0])
	num, _ := strconv.Atoi(position[1:])

	pos1 := fmt.Sprintf("%s%d", letter, num-1)
	pos2 := fmt.Sprintf("%s%d", letter, num+1)

	return pos1, pos2
}

func get_squares_along_row(pos string) (string, string) {

	letter := string(pos[0])
	num, _ := strconv.Atoi(pos[1:])

	letter_index := boardIndex[letter]
	letter_before_idx := letter_index - 1
	letter_after_idx := letter_index + 1

	letter_before := boardLetters[letter_before_idx]
	letter_after := boardLetters[letter_after_idx]

	pos_before := fmt.Sprintf("%s%d", letter_before, num+1)
	pos_after := fmt.Sprintf("%s%d", letter_after, num+1)

	return pos_before, pos_after

}

func (k Knight) GetColor() string {
    return k.Color
}

func (k Knight) GetPosition() string {
    return k.Position
}

func (k Knight) GetPieceType() string {
    return k.PieceType
}