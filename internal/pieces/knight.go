package pieces

import (

	"github.com/Glenn444/golang-chess/utils"
)

type Knight struct {
	PieceType string
	Color     string
	Position  string
}

// var boardIndex = map[string]int{
// 	"a": 0,
// 	"b": 1,
// 	"c": 2,
// 	"d": 3,
// 	"e": 4,
// 	"f": 5,
// 	"g": 6,
// 	"h": 7,
// }
// var boardLetters = []string{"a", "b", "c", "d", "e", "f", "g", "h"}

// current_position = "a2"
func (k Knight) GetLegalSquares() []string {

	var possible_positions []string
	//letter := string(k.Position[0])
	//num, _ := strconv.Atoi(k.Position[1:])

	row, col := utils.Chess_notation_to_indices(k.Position)

	knightMoves := [][2]int{{2,1}, {2,-1}, {-2,1}, {-2,-1}, {1,2}, {1,-2}, {-1,2}, {-1,-2}}

	for _, k_move := range knightMoves{
		new_row := row + k_move[0]
		new_col := col + k_move[1]

		if new_row >= 0 && new_row < 8 && new_col >= 0 && new_col < 8{
		pos := utils.Indices_to_chess_notation(new_row,new_col)
		possible_positions = append(possible_positions, pos)
		}
	}

	return possible_positions
}

// func get_squares_along_column(position string) (string, string) {
// 	letter := string(position[0])
// 	num, _ := strconv.Atoi(position[1:])

// 	if num < 0 || num > 7{
// 		return  "",""
// 	}

// 	pos1 := fmt.Sprintf("%s%d", letter, num-1)
// 	pos2 := fmt.Sprintf("%s%d", letter, num+1)

// 	return pos1, pos2
// }

// func get_squares_along_row(pos string) (string, string) {

// 	letter := string(pos[0])
// 	num, _ := strconv.Atoi(pos[1:])
// 	if num < 0 || num > 7{
// 		return  "",""
// 	}

// 	letter_index := boardIndex[letter]
// 	letter_before_idx := letter_index - 1
// 	letter_after_idx := letter_index + 1

// 	letter_before := boardLetters[letter_before_idx]
// 	letter_after := boardLetters[letter_after_idx]

// 	pos_before := fmt.Sprintf("%s%d", letter_before, num+1)
// 	pos_after := fmt.Sprintf("%s%d", letter_after, num+1)

// 	return pos_before, pos_after

// }

func (k Knight) GetColor() string {
	return k.Color
}

func (k Knight) GetPosition() string {
	return k.Position
}

func (k Knight) GetPieceType() string {
	return k.PieceType
}
