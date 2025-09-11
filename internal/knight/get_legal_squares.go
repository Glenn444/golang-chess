package knight

import (
	"fmt"
	"strconv"

	//"github.com/Glenn444/golang-chess/internal/board"
	"github.com/Glenn444/golang-chess/utils"
)

// type knight struct {
// 	board.Piece
// }

// current_position = "a2"
func Get_legal_squares(current_postion string) []string {

	boardLetters := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	var possible_positions []string
	letter := string(current_postion[0])
	num, _ := strconv.Atoi(current_postion[1:])

	row, col := utils.Chess_notation_to_indices(current_postion)

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
