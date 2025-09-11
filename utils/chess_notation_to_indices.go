package utils

import (
	"strconv"
)

// position can be a1 -> board[0][0]
// a1 -> a: column position, 1: row position
func Chess_notation_to_indices(position string) (int, int) {

	letter := string(position[0])
	num, _ := strconv.Atoi(position[1:])
	var row_pos, col_pos int

	board_letters := []string{"a", "b", "c", "d", "e", "f", "g", "h"}

	for k, v := range board_letters {
		if v == letter {
			row_pos = num - 1
			col_pos = k
		}
	}

	return row_pos, col_pos

}
