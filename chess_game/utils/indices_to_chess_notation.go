package utils

import "fmt"

func Indices_to_chess_notation(row_pos,col_pos int) string  {
	
	board_letters := []string{"a","b","c","d","e","f","g","h"}

	var position string
	var letter string

	for k,v:= range board_letters{
		if k == col_pos {
			letter = v
		}
	}
	//0,0 -> a1

	position = fmt.Sprintf("%s%d",letter,row_pos+1)
	

	return  position
}