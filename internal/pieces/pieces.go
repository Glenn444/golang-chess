package pieces



type PieceInterface interface{
	GetLegalSquares() []string
}