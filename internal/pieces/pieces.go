package pieces



type PieceInterface interface{
	GetLegalSquares() []string
	GetColor() string
    GetPosition() string
    GetPieceType() string
	AssignPosition(pos string)
	String() string
}