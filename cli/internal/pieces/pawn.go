package pieces

import (
	"fmt"
	"strconv"

	"github.com/Glenn444/golang-chess/utils"
)

type Pawn struct {
	PieceType string
	Color     string
	Position  string
	Points    int64
}

func (p *Pawn) GetLegalSquares(g GameState) []string {
	var positions []string

	letter := string(p.Position[0])
	num, _ := strconv.Atoi(p.Position[1:])
	var added bool

	switch num {
	case 2:
		added = true
		pos1 := fmt.Sprintf("%s%d", letter, num+1)
		rowpos1, colpos1 := utils.Chess_notation_to_indices(pos1)
		if isInBounds(rowpos1, colpos1) && !g.Board[rowpos1][colpos1].Occupied {
			positions = append(positions, pos1)

			pos2 := fmt.Sprintf("%s%d", letter, num+2)
			rowpos2, colpos2 := utils.Chess_notation_to_indices(pos2)

			if isInBounds(rowpos2, colpos2) && !g.Board[rowpos2][colpos2].Occupied {
				positions = append(positions, pos2)
			}
		}

		//positions = append(positions, pos1, pos2)

		row, col := utils.Chess_notation_to_indices(p.Position)
		diagRowR := row + 1
		diagColR := col + 1
		if isInBounds(diagRowR, diagColR) && g.Board[diagRowR][diagColR].Occupied && g.Board[diagRowR][diagColR].Piece.GetColor() != p.Color {
			diagRight := utils.Indices_to_chess_notation(diagRowR, diagColR)
			positions = append(positions, diagRight)
		}

		diagRowL := row + 1
		diagColL := col - 1
		if isInBounds(diagRowL, diagColL) && g.Board[diagRowL][diagColL].Occupied && g.Board[diagRowL][diagColL].Piece.GetColor() != p.Color {
			diagRight := utils.Indices_to_chess_notation(diagRowL, diagColL)
			positions = append(positions, diagRight)
		}

	case 7:
		added = true
		pos1 := fmt.Sprintf("%s%d", letter, num-1)
		rowpos1, colpos1 := utils.Chess_notation_to_indices(pos1)
		if !g.Board[rowpos1][colpos1].Occupied {
			positions = append(positions, pos1)

			pos2 := fmt.Sprintf("%s%d", letter, num-2)
			rowpos2, colpos2 := utils.Chess_notation_to_indices(pos2)

			if !g.Board[rowpos2][colpos2].Occupied {
				positions = append(positions, pos2)
			}
		}
		//positions = append(positions, pos1, pos2)

		row, col := utils.Chess_notation_to_indices(p.Position)

		diagRowR := row - 1
		diagColR := col - 1

		if isInBounds(diagRowR, diagColR) && g.Board[diagRowR][diagColR].Occupied && g.Board[diagRowR][diagColR].Piece.GetColor() != p.Color {
			diagRight := utils.Indices_to_chess_notation(diagRowR, diagColR)
			positions = append(positions, diagRight)
		}

		diagRowL := row - 1
		diagColL := col + 1

		if isInBounds(diagRowL, diagColL) && g.Board[diagRowL][diagColL].Occupied && g.Board[diagRowL][diagColL].Piece.GetColor() != p.Color {
			diagRight := utils.Indices_to_chess_notation(diagRowL, diagColL)
			positions = append(positions, diagRight)
		}
	}

	if !added {
		switch p.Color {
		case "w":
			pos1 := fmt.Sprintf("%s%d", letter, num+1)
			rowpos1, colpos1 := utils.Chess_notation_to_indices(pos1)
			if isInBounds(rowpos1, colpos1) && g.Board[rowpos1][colpos1].Occupied && g.Board[rowpos1][colpos1].Piece.GetColor() != p.Color {
				positions = append(positions, pos1)
			} else if isInBounds(rowpos1, colpos1) && !g.Board[rowpos1][colpos1].Occupied {
				positions = append(positions, pos1)
			}
			//positions = append(positions, pos1)

			row, col := utils.Chess_notation_to_indices(p.Position)
			diagRowR := row + 1
			diagColR := col + 1
			if isInBounds(diagRowR, diagColR) && g.Board[diagRowR][diagColR].Occupied && g.Board[diagRowR][diagColR].Piece.GetColor() != p.Color {
				diagRight := utils.Indices_to_chess_notation(diagRowR, diagColR)
				positions = append(positions, diagRight)
			}

			diagRowL := row + 1
			diagColL := col - 1
			if isInBounds(diagRowL, diagColL) && g.Board[diagRowL][diagColL].Occupied && g.Board[diagRowL][diagColL].Piece.GetColor() != p.Color {
				diagRight := utils.Indices_to_chess_notation(diagRowL, diagColL)
				positions = append(positions, diagRight)
			}
		case "b":
			pos1 := fmt.Sprintf("%s%d", letter, num-1)
			rowpos1, colpos1 := utils.Chess_notation_to_indices(pos1)
			if isInBounds(rowpos1, colpos1) && g.Board[rowpos1][colpos1].Occupied && g.Board[rowpos1][colpos1].Piece.GetColor() != p.Color {
				positions = append(positions, pos1)
			} else if !g.Board[rowpos1][colpos1].Occupied {
				positions = append(positions, pos1)
			}
			//positions = append(positions, pos1)

			row, col := utils.Chess_notation_to_indices(p.Position)
			diagRowR := row - 1
			diagColR := col - 1
			if isInBounds(diagRowR, diagColR) && g.Board[diagRowR][diagColR].Occupied && g.Board[diagRowR][diagColR].Piece.GetColor() != p.Color {
				diagRight := utils.Indices_to_chess_notation(diagRowR, diagColR)
				positions = append(positions, diagRight)
			}

			diagRowL := row - 1
			diagColL := col + 1
			if isInBounds(diagRowL, diagColL) && g.Board[diagRowL][diagColL].Occupied && g.Board[diagRowL][diagColL].Piece.GetColor() != p.Color {
				diagRight := utils.Indices_to_chess_notation(diagRowL, diagColL)
				positions = append(positions, diagRight)
			}
		}
	}
	return positions
}

func isInBounds(row, col int) bool {
	return row >= 0 && row < 8 && col >= 0 && col < 8
}

func (p *Pawn) GetColor() string {
	return p.Color
}

func (p *Pawn) GetPosition() string {
	return p.Position
}

func (p *Pawn) GetPieceType() string {
	return p.PieceType
}

func (p *Pawn) AssignPosition(pos string) {
	p.Position = pos
}

func (p *Pawn)Clone()PieceInterface{
	 
	return &Pawn{
		Color: p.Color,
		PieceType: p.PieceType,
		Position: p.Position,
	}
}
func (p *Pawn) String() string {
	if p.Color == "w" {
		return "[♙]" // or "wP"
	}
	return "[♟]" // or "bP"
}
func (p *Pawn) GetPiecePoints() int64 {
	return p.Points
}
