package knight

import (
	"fmt"
	"strconv"
)

func get_squares_along_row(pos string) (string, string) {
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
	boardLetters := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	letter := string(pos[0])
	num, _ := strconv.Atoi(pos[1:])

	letter_index := boardIndex[letter]
	letter_before_idx := letter_index - 1
	letter_after_idx := letter_index + 1

	letter_before := boardLetters[letter_before_idx]
	letter_after := boardLetters[letter_after_idx]

	pos_before := fmt.Sprintf("%s%d",letter_before,num+1)
	pos_after := fmt.Sprintf("%s%d",letter_after,num+1)

	return  pos_before,pos_after

}
