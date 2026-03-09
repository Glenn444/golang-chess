package pieces

type Square struct {
	Occupied bool
	Piece    PieceInterface
}

type GameState struct {
	CurrentPlayer string
	Board         [][]Square
}

type PieceInterface interface{
	GetLegalSquares(g GameState) []string
	GetColor() string
    GetPosition() string
    GetPieceType() string
	AssignPosition(pos string)
	String() string
}