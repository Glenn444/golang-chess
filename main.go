package main

import (
	"fmt"

	"github.com/Glenn444/golang-chess/internal/knight"
	"github.com/Glenn444/golang-chess/utils"
)



func main() {

	//board_letters := []string{"a","b","c","d","e","f","g","h"}
	//board := board.Print_board()

	// fmt.Printf("    ")
	// for _,c := range "abcdefgh"{
	// 	fmt.Printf("%c ",c)
	// }
	// fmt.Printf("\n")
	// for i,row := range board{
	// 	fmt.Printf("%d, %v\n",i,row)
	// 	fmt.Printf("    ")
	// 	for j := range row{
	// 		//positon := fmt.Sprintf("%s%d",board_letters[j],i+1)
	// 		fmt.Printf("%s%d ",board_letters[j],i+1)
			
	// 	}
	// 	fmt.Printf("\n")
		
	// }

	pos := utils.Indices_to_chess_notation(4,3)
	row_pos,col_pos := utils.Chess_notation_to_indices("d5")
	possible_positions := knight.Get_legal_squares("f4")
	
	fmt.Printf("position: %s\n",pos)
	fmt.Printf("row: %d, col: %d\n",row_pos,col_pos)
	fmt.Printf("knight squares: %v\n",possible_positions)
}