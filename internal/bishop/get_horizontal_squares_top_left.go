package bishop

import (
	"fmt"

	"github.com/Glenn444/golang-chess/utils"
)

func get_horizontal_squares_top_left(pos string) []string  {
	board_letters := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	nums := []int{1,2,3,4,5,6,7,8}

	row,col := utils.Chess_notation_to_indices(pos)
	diagnol := row-col
	var possible_possitions []string

	for _,v := range board_letters{
		for _,j := range nums{
			position := fmt.Sprintf("%s%d",v,j)
			row,col := utils.Chess_notation_to_indices(position)
			diag := row - col

			if diagnol == diag{
				possible_possitions = append(possible_possitions, position)
			}
		}
	}

	return  possible_possitions
}