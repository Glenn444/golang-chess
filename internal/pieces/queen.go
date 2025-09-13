package pieces

type Queen struct {
	Color     string
	PieceType string
	Position  string
}

func (q Queen) GetLegalSquares() []string {
	b := Bishop{
		Color:     q.Color,
		PieceType: "B",
		Position:  q.Position,
	}

	r := Rook{
		Color:     q.Color,
		PieceType: "R",
		Position:  q.Position,
	}
	var positions []string
	rook_pos := r.GetLegalSquares()
	bishop_pos := b.GetLegalSquares()

	positions = append(positions, rook_pos...)
	positions = append(positions, bishop_pos...)

	return positions
}
