package pieces

import (
	"fmt"

	"github.com/Glenn444/golang-chess/internal/utils/chess"
)

type King struct {
	PieceType string
	Color     string
	Position string
	Points 	int64
}

func (k *King) GetLegalSquares(g GameState) []string {
	// board files (columns) and ranks (rows)
	files := "abcdefgh"
	rank := int(k.Position[1] - '0')
	file := int(k.Position[0] - 'a')

	// directions: 8 possible moves (dx, dy)
	directions := [][2]int{
		{1, 0}, {-1, 0}, // left, right
		{0, 1}, {0, -1}, // up, down
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1}, // diagonals
	}

	var moves []string
	for _, d := range directions {
		newFile := file + d[0]
		newRank := rank + d[1]

		// check bounds: file 0–7, rank 1–8
		if newFile >= 0 && newFile < 8 && newRank >= 1 && newRank <= 8 {
			move := string(files[newFile]) + fmt.Sprint(newRank)
			row,col, _ := chess.ChessNotationToIndices(move)
			if g.Board[row][col].Occupied && g.Board[row][col].Piece.GetColor() != k.Color{
			moves = append(moves, move)
			}else if !g.Board[row][col].Occupied{
				moves = append(moves, move)
			}
		}
	}

	return moves

}

func (k *King) GetColor() string {
    return k.Color
}

func (k *King) GetPosition() string {
    return k.Position
}

func (k *King) GetPieceType() string {
    return k.PieceType
}

func (k *King) AssignPosition(pos string){
	k.Position = pos
}
func (k *King) GetPiecePoints()int64{
	return k.Points
}

func (k *King)Clone()PieceInterface{
	 
	return &King{
		Color: k.Color,
		PieceType: k.PieceType,
		Position: k.Position,
	}
}
func (k *King) String() string {
    if k.Color == "w" {
        return "[♔]" // or "wK"
    }
    return "[♚]" // or "bK"
}