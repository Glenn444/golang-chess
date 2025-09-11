package knight

import (
	"fmt"
	"strconv"
)

func get_squares_along_column(position string) (string,string)  {
	letter := string(position[0])
	num,_ := strconv.Atoi(position[1:])

	pos1 := fmt.Sprintf("%s%d",letter,num-1)
	pos2 := fmt.Sprintf("%s%d",letter,num+1)

	return pos1,pos2
}