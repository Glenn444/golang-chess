package rook

import (
	"fmt"
	"strconv"
)

func Get_legal_squares(pos string) []string  {
	boardIndex := map[string]int{
		"a": 0,
		"b": 1,
		"c": 2,
		"d": 3,
		"e": 4,
		"f": 5,
		"g": 6,
		"h": 7,
	}
	var possible_possitions []string

	letter := string(pos[0])
	num,_ := strconv.Atoi(pos[1:])

	var column_pos []string
	var row_pos []string

	for k,v := range boardIndex{
		pos_c := fmt.Sprintf("%s%d",letter,v+1)
		column_pos = append(column_pos, pos_c)

		pos_r := fmt.Sprintf("%s%d",k,num)
		row_pos = append(row_pos, pos_r)
	}

	possible_possitions = append(possible_possitions, column_pos...)
	possible_possitions = append(possible_possitions, row_pos...)

	return possible_possitions
}