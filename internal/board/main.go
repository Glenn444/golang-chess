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



func Print_board() [][]int  {
	rows,cols := 8,8

	board := make([][]int,rows)

	for i := range board{
		board[i] = make([]int, cols)
	}

	return board
}