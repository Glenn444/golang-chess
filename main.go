package main

import (
	"fmt"

	"github.com/Glenn444/golang-chess/internal/board"
	
)

type GameState struct{
	CurrentPlayer string
	GameOn bool
}

func main() {

	initialBoard := board.Create_board()

	b := board.Initialise_board(initialBoard)
	for i,row := range b{
		fmt.Printf("%d, %v\n",i,row)
		fmt.Printf("    ")
		for _,v := range row{
			//positon := fmt.Sprintf("%s%d",board_letters[j],i+1)
			fmt.Printf("%v ",v)
			
		}
		fmt.Printf("\n")
		
	}






	//pos := utils.Indices_to_chess_notation(0,0)//"a1"
	//row_pos,col_pos := utils.Chess_notation_to_indices("d5")
	//possible_positions_knight := knight.Get_legal_squares("f4")
	//possible_positions_rook := rook.Get_legal_squares("h5")
	//possible_positions_bishop := bishop.Get_legal_squares("f1")
	//possible_positions_queen := queen.Get_legal_squares("d1")

	
	// fmt.Printf("position: %s\n",pos)
	// fmt.Printf("row: %d, col: %d\n",row_pos,col_pos)
	// //fmt.Printf("knight squares: %v\n",possible_positions_knight)
	// fmt.Printf("rook squares: %v\n",possible_positions_rook)
	// fmt.Printf("bishop squares: %v\n",possible_positions_bishop)
	// fmt.Printf("Queen squares: %v\n",possible_positions_queen)
	
}