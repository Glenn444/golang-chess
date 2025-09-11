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

	board_letters := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	var possible_positions []string
	//letter := string(current_postion[0])
	num, _ := strconv.Atoi(current_postion[1:])

	row, col := utils.Chess_notation_to_indices(current_postion)

	row_top := row + 2
	row_bottom := row - 2
	col_left := col - 2
	col_right := col + 2

	col_letter_left := col - 1
	col_letter_right := col + 1

	

	for k, v := range board_letters {
		var col_right_key int
		switch k {
		
		case col_left:
			pos1 := fmt.Sprintf("%s%d", v, num+1)
			pos2 := fmt.Sprintf("%s%d", v, num-1)
			possible_positions = append(possible_positions, pos1)
			possible_positions = append(possible_positions, pos2)
		case col_right:
			pos := fmt.Sprintf("%s%d", v, num+1)
			possible_positions = append(possible_positions, pos)
			col_right_key = k
		case col_letter_left:
			pos := fmt.Sprintf("%s%d", v, row_bottom+1)
			possible_positions = append(possible_positions, pos)
		case col_letter_right:
			pos := fmt.Sprintf("%s%d", v, row_top+1)
			possible_positions = append(possible_positions, pos)
		case col_right_key + 2:
			pos := fmt.Sprintf("%s%d", v, num)
			possible_positions = append(possible_positions, pos)
		}


	}

	return possible_positions
}

//rewrite into get column squares
//rewrite into get row squares