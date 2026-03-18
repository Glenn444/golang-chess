package utils

type Square struct {
	Occupied bool
	Piece    PieceInterface
}

type PieceInterface interface{
	GetLegalSquares() []string
	GetColor() string
    GetPosition() string
    GetPieceType() string
	AssignPosition(pos string)
	String() string
	Clone() PieceInterface
}
func CopySlice(copyA[][]Square,copyB[][]Square){
	//copy b to A
	for i,squares := range copyB{
		for j,square := range squares{
			copyA[i][j] = Square{
				Occupied: square.Occupied,
				
			}
		}
	}
}