package pieces

import (
	"fmt"
	"strconv"
)

type Pawn struct {
	Color     string
	PieceType string
	Position  string
}

func (p Pawn) GetLegalSquares() []string {
	var positions []string
	var initialPos = []string{
		"a2", "b2", "c2", "d2", "e2", "f2", "g2", "h2", // white pawns
		"a7", "b7", "c7", "d7", "e7", "f7", "g7", "h7", // black pawns
	}

	letter := string(p.Position[0])
	num, _ := strconv.Atoi(p.Position[1:])
	

	for _,pos := range initialPos {
		if p.Position == pos &&  p.Color == "w"{
			pos1 := fmt.Sprintf("%s%d", letter, num+1)
			pos2 := fmt.Sprintf("%s%d", letter, num+2)
			positions = append(positions,pos1,pos2)
		}else if p.Position == pos &&  p.Color == "b"{
			pos1 := fmt.Sprintf("%s%d", letter, num-1)
			pos2 := fmt.Sprintf("%s%d", letter, num-2)
			positions = append(positions,pos1,pos2)
		}else if !(p.Position == pos) && p.Color == "w"{
			pos1 := fmt.Sprintf("%s%d", letter, num+1)
			positions = append(positions,pos1,pos1)
		}else if !(p.Position == pos) && p.Color == "b"{
			pos1 := fmt.Sprintf("%s%d", letter, num-1)
			positions = append(positions,pos1,pos1)
		}
	}

	return positions
}
