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
}

func (p Pawn) GetLegalSquares(g GameState) []string {
	var positions []string

	letter := string(p.Position[0])
	num, _ := strconv.Atoi(p.Position[1:])
	var added bool

	switch num {
	case 2:
		added = true
		pos1 := fmt.Sprintf("%s%d", letter, num+1)
		pos2 := fmt.Sprintf("%s%d", letter, num+2)
		positions = append(positions, pos1, pos2)

		row, col := utils.Chess_notation_to_indices(p.Position)
		diagRowR := row + 1
		diagColR := col + 1
		if g.Board[diagRowR][diagColR].Occupied && g.Board[diagRowR][diagColR].Piece.GetColor() != p.Color {
			diagRight := utils.Indices_to_chess_notation(diagRowR, diagColR)
			positions = append(positions, diagRight)
		}

		diagRowL := row + 1
		diagColL := col - 1
		if g.Board[diagRowL][diagColL].Occupied && g.Board[diagRowL][diagColL].Piece.GetColor() != p.Color {
			diagRight := utils.Indices_to_chess_notation(diagRowL, diagColL)
			positions = append(positions, diagRight)
		}

	case 7:
		added = true
		pos1 := fmt.Sprintf("%s%d", letter, num-1)
		pos2 := fmt.Sprintf("%s%d", letter, num-2)
		positions = append(positions, pos1, pos2)

		row, col := utils.Chess_notation_to_indices(p.Position)
		diagRowR := row - 1
		diagColR := col - 1
		if g.Board[diagRowR][diagColR].Occupied && g.Board[diagRowR][diagColR].Piece.GetColor() != p.Color {
			diagRight := utils.Indices_to_chess_notation(diagRowR, diagColR)
			positions = append(positions, diagRight)
		}

		diagRowL := row - 1
		diagColL := col + 1
		if g.Board[diagRowL][diagColL].Occupied && g.Board[diagRowL][diagColL].Piece.GetColor() != p.Color {
			diagRight := utils.Indices_to_chess_notation(diagRowL, diagColL)
			positions = append(positions, diagRight)
		}
	}

	if !added {
		switch p.Color {
		case "w":
			pos1 := fmt.Sprintf("%s%d", letter, num+1)
			positions = append(positions, pos1)

			row, col := utils.Chess_notation_to_indices(p.Position)
			diagRowR := row + 1
			diagColR := col + 1
			if g.Board[diagRowR][diagColR].Occupied && g.Board[diagRowR][diagColR].Piece.GetColor() != p.Color {
				diagRight := utils.Indices_to_chess_notation(diagRowR, diagColR)
				positions = append(positions, diagRight)
			}

			diagRowL := row + 1
			diagColL := col - 1
			if g.Board[diagRowL][diagColL].Occupied && g.Board[diagRowL][diagColL].Piece.GetColor() != p.Color {
				diagRight := utils.Indices_to_chess_notation(diagRowL, diagColL)
				positions = append(positions, diagRight)
			}
		case "b":
			pos1 := fmt.Sprintf("%s%d", letter, num-1)
			positions = append(positions, pos1)

			row, col := utils.Chess_notation_to_indices(p.Position)
			diagRowR := row - 1
			diagColR := col - 1
			if g.Board[diagRowR][diagColR].Occupied && g.Board[diagRowR][diagColR].Piece.GetColor() != p.Color {
				diagRight := utils.Indices_to_chess_notation(diagRowR, diagColR)
				positions = append(positions, diagRight)
			}

			diagRowL := row - 1
			diagColL := col + 1
			if g.Board[diagRowL][diagColL].Occupied && g.Board[diagRowL][diagColL].Piece.GetColor() != p.Color {
				diagRight := utils.Indices_to_chess_notation(diagRowL, diagColL)
				positions = append(positions, diagRight)
			}
		}
	}
	return positions
}

func (p Pawn) GetColor() string {
	return p.Color
}

func (p Pawn) GetPosition() string {
	return p.Position
}

func (p Pawn) GetPieceType() string {
	return p.PieceType
}

func (p *Pawn) AssignPosition(pos string) {
	p.Position = pos
}

func (p Pawn) String() string {
	if p.Color == "w" {
		return "[♙]" // or "wP"
	}
	return "[♟]" // or "bP"
}
