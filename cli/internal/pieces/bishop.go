package pieces

import (
	"fmt"

	"github.com/Glenn444/golang-chess/utils"
)


// var board_letters = []string{"a", "b", "c", "d", "e", "f", "g", "h"}
// var nums = []int{1, 2, 3, 4, 5, 6, 7, 8}



type Bishop struct {
	PieceType string
	Color     string
	Position  string
}

func (b Bishop) GetLegalSquares(g GameState) []string {
	var positions []string
	//pos_top_left := get_horizontal_squares_top_left(b.Position)
	//pos_top_right := get_horizontal_squares_top_right(b.Position)

	//positions = append(positions, pos_top_right...)
	//positions = append(positions, pos_top_left...)

	fmt.Printf("diagnol squares top right d4: %v\n",getDiagnolSquares("d4",1,1))
	allDiagnols := [][]string{
		getDiagnolSquares(b.Position,1,1),
		getDiagnolSquares(b.Position,1,-1),
		getDiagnolSquares(b.Position,-1,1),
		getDiagnolSquares(b.Position,-1,-1),
	}
	//return positions
	for _,diagnol := range allDiagnols{
		for _, pos := range diagnol{
			i, j := utils.Chess_notation_to_indices(pos)
			square := g.Board[i][j]

			if square.Occupied{
				if square.Piece.GetColor() != b.Color{
					positions = append(positions, pos)
				}
				break
			}
			
			positions = append(positions, pos)
		}
		

	}
	//fmt.Printf("legalsquares: %v\n",positions)
	return positions
}


// func get_horizontal_squares_top_left(pos string) []string {
// 	var diagnol int
// 	row, col := utils.Chess_notation_to_indices(pos)
// 	//fmt.Printf("row: %d, col: %d\n", row, col)


// 	var possible_possitions []string

// 	if row >= 0 && row < 8 && col >= 0 && col < 8 {
// 		diagnol = row - col
// 	}

// 	for _, v := range board_letters {
// 		for _, j := range nums {
// 			position := fmt.Sprintf("%s%d", v, j)
// 			row1, col1 := utils.Chess_notation_to_indices(position)
// 			diag := row1 - col1

// 			if diagnol == diag && pos != position {
// 				possible_possitions = append(possible_possitions, position)
// 			}
// 		}
// 	}

// 	return possible_possitions
// }

// func get_horizontal_squares_top_right(pos string) []string {
// 	var diagnol int
// 	row, col := utils.Chess_notation_to_indices(pos)
// 	//dia := row + col
// 	var possible_possitions []string

// 	if row >= 0 && row < 8 && col >= 0 && col < 8 {
// 		diagnol = row + col
// 	}
// 	for _, v := range board_letters {
// 		for _, j := range nums {
// 			position := fmt.Sprintf("%s%d", v, j)
// 			row1, col1 := utils.Chess_notation_to_indices(position)
// 			diag := row1 + col1

// 			if diagnol == diag && pos != position{
// 				possible_possitions = append(possible_possitions, position)
// 			}
// 		}
// 	}

// 	return possible_possitions
// }

func getDiagnolSquares(pos string,rowDelta int, colDelta int)[]string{
	var positions []string
	row,col := utils.Chess_notation_to_indices(pos)

	
	for true{
		r := row + rowDelta
		c := col + colDelta
		if r >= 8 || r < 0 || c >= 8 || c < 0{
			break
		}
		
		pos := utils.Indices_to_chess_notation(r,c)
		positions = append(positions, pos)
		row = r
		col = c
	}
	return positions
}

func (b Bishop) GetColor() string {
	return b.Color
}

func (b Bishop) GetPosition() string {
	return b.Position
}

func (b Bishop) GetPieceType() string {
	return b.PieceType
}

func (b *Bishop) AssignPosition(pos string) {
	b.Position = pos
}

func (b Bishop) String() string {
	if b.Color == "w" {
		return "[♗]" // or "wB" if you prefer text
	}
	return "[♝]" // or "bB" if you prefer text
}
