package board



type Square struct{
Color string
Occupied bool
Piece Piece
}

type Piece struct{
Color string
PieceType string
Position string
}

type Piecetype struct{

}



func Print_board() [][]Square  {
	rows,cols := 8,8

	board := make([][]Square,rows)

	for i := range board{
		board[i] = make([]Square, cols)
	}

	return board
}