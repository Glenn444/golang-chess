package pieces

import (

	"github.com/Glenn444/golang-chess/internal/utils/chess"
)


type Bishop struct {
	PieceType string
	Color     string
	Position  string
	Points	int64
}

func (b *Bishop) GetLegalSquares(g GameState) []string {
	var positions []string

	//fmt.Printf("diagnol squares top right d4: %v\n",getDiagnolSquares("d4",1,1))
	allDiagnols := [][]string{
		getDiagnolSquares(b.Position,1,1),
		getDiagnolSquares(b.Position,1,-1),
		getDiagnolSquares(b.Position,-1,1),
		getDiagnolSquares(b.Position,-1,-1),
	}
	//return positions
	for _,diagnol := range allDiagnols{
		for _, pos := range diagnol{
			i, j := chess.Chess_notation_to_indices(pos)
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


func getDiagnolSquares(pos string,rowDelta int, colDelta int)[]string{
	var positions []string
	row,col := chess.Chess_notation_to_indices(pos)

	
	for true{
		r := row + rowDelta
		c := col + colDelta
		if r >= 8 || r < 0 || c >= 8 || c < 0{
			break
		}
		
		pos := chess.Indices_to_chess_notation(r,c)
		positions = append(positions, pos)
		row = r
		col = c
	}
	return positions
}

func (b *Bishop) GetColor() string {
	return b.Color
}

func (b *Bishop) GetPosition() string {
	return b.Position
}

func (b *Bishop) GetPieceType() string {
	return b.PieceType
}

func (b *Bishop) AssignPosition(pos string) {
	b.Position = pos
}

func (b *Bishop)Clone()PieceInterface{
	 
	return &Bishop{
		Color: b.Color,
		PieceType: b.PieceType,
		Position: b.Position,
		Points: b.Points,
	}
}

func (b *Bishop) String() string {
	if b.Color == "w" {
		return "[♗]" // or "wB" if you prefer text
	}
	return "[♝]" // or "bB" if you prefer text
}

func (b *Bishop) GetPiecePoints()int64{
	return b.Points
}
