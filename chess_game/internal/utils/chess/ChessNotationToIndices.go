package chess

import (
	
	"strconv"
)


// position can be a1 -> board[0][0]
// a1 -> a: column position, 1: row position
func ChessNotationToIndices(position string) (int, int,error) {

	posLen := len(position)
	if posLen != 2{
		return 0,0,ErrInvalidPos
	}
	letter := position[0]
	num, err := strconv.Atoi(position[1:])
	if err != nil{
		return 0,0,ErrInvalidPos
	}
	var rowPos, colPos int

	boardLetters := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}

	for k, v := range boardLetters {
		if v == letter {
			rowPos = num - 1
			colPos = k
			return rowPos,colPos,nil
		}
	}

	return 0,0,ErrInvalidPos

}
