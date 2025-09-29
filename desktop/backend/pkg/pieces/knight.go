package pieces

import (

	"github.com/Glenn444/golang-chess/backend/utils"
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


func (k Knight) GetColor() string {
	return k.Color
}

func (k Knight) GetPosition() string {
	return k.Position
}

func (k Knight) GetPieceType() string {
	return k.PieceType
}

func (k *Knight) AssignPosition(pos string){
	k.Position = pos
}

func (k Knight) String() string {
    if k.Color == "w" {
        return "[♘]" // or "wN"
    }
    return "[♞]" // or "bN"
}