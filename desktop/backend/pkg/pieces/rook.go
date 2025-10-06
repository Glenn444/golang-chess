package pieces

import (
	"fmt"
	"strconv"
)

type Rook struct{
	PieceType string
	Color string
	Position string
}

func (r Rook) GetLegalSquares() []string  {
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

	letter := string(r.Position[0])
	num,_ := strconv.Atoi(r.Position[1:])

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



func (r Rook) GetColor() string {
    return r.Color
}

func (r Rook) GetPosition() string {
    return r.Position
}

func (r Rook) GetPieceType() string {
    return r.PieceType
}

func (r *Rook) AssignPosition(pos string){
	r.Position = pos
}

func (r Rook) String() string {
    if r.Color == "w" {
        return "wR" // or "wR"
    }
    return "bR" // or "bR"
}