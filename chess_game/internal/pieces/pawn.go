package pieces

import (
	"github.com/Glenn444/golang-chess/internal/utils/chess"
)

type Pawn struct {
	PieceType string
	Color     string
	Position  string
	Points    int64
}

func (p *Pawn) GetLegalSquares(g *GameState) []string {
	var positions []string

	row, col, err := chess.ChessNotationToIndices(p.Position)
	if err != nil {
		return positions
	}

	// White pawns move up the board, black pawns move down.
	dir := 1
	homeRow := 1
	if p.Color == "b" {
		dir = -1
		homeRow = 6
	}

	// Single push, and double push from the home rank.
	pushRow := row + dir
	if isInBounds(pushRow, col) && !g.Board[pushRow][col].Occupied {
		positions = append(positions, chess.Indices_to_chess_notation(pushRow, col))

		doubleRow := row + 2*dir
		if row == homeRow && isInBounds(doubleRow, col) && !g.Board[doubleRow][col].Occupied {
			positions = append(positions, chess.Indices_to_chess_notation(doubleRow, col))
		}
	}

	// Diagonal captures, including en passant onto the empty target square.
	for _, dc := range []int{-1, 1} {
		capRow, capCol := row+dir, col+dc
		if !isInBounds(capRow, capCol) {
			continue
		}
		target := chess.Indices_to_chess_notation(capRow, capCol)
		square := g.Board[capRow][capCol]
		if square.Occupied && square.Piece.GetColor() != p.Color {
			positions = append(positions, target)
		} else if !square.Occupied && g.EnPassantTarget != "" && target == g.EnPassantTarget {
			positions = append(positions, target)
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
		Points: p.Points,
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
