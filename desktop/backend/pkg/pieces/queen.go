package pieces

type Queen struct {
	PieceType string
	Color     string
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

func (q Queen) GetColor() string {
	return q.Color
}

func (q Queen) GetPosition() string {
	return q.Position
}

func (q Queen) GetPieceType() string {
	return q.PieceType
}

func (q *Queen) AssignPosition(pos string) {
	q.Position = pos
}

func (q Queen) String() string {
	if q.Color == "white" {
		return "wQ" // or "wQ"
	}
	return "bQ" // or "bQ"
}
