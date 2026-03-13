package pieces

import (

	"github.com/Glenn444/golang-chess/utils"
)

type Rook struct {
	PieceType string
	Color     string
	Position  string
}

func (r Rook) GetLegalSquares(g GameState) []string {
	var positions []string

	allDiagnols := [][]string{
		getHorizontalVertical(r,-1,0),
		getHorizontalVertical(r,0,-1),
		getHorizontalVertical(r,0,1),
		getHorizontalVertical(r,1,0),
	}
	//return positions
	for _,diagnol := range allDiagnols{
		for _, pos := range diagnol{
			i, j := utils.Chess_notation_to_indices(pos)
			square := g.Board[i][j]

			if square.Occupied{
				if square.Piece.GetColor() != r.Color{
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

func getHorizontalVertical(r Rook,rowDelta int, colDelta int) []string {
	
	var possible_positions []string

	row,col := utils.Chess_notation_to_indices(r.Position)
	
	for true{
		r := row+rowDelta
		c := col+colDelta

		if r >= 8 || r < 0 || c >= 8 || c < 0{
			break
		}
		pos := utils.Indices_to_chess_notation(r,c)
		possible_positions = append(possible_positions, pos)
		row = r
		col = c
	}
	return possible_positions
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

func (r *Rook) AssignPosition(pos string) {
	r.Position = pos
}

func (r Rook) String() string {
	if r.Color == "w" {
		return "[♖]" // or "wR"
	}
	return "[♜]" // or "bR"
}
