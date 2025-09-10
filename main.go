package main

import (
	"fmt"

	"github.com/Glenn444/golang-chess/internal/board"
)



func main() {
	board_letters := []string{"a","b","c","d","e","f","g","h"}
	board := board.Print_board()

	fmt.Printf("    ")
	for _,c := range "abcdefgh"{
		fmt.Printf("%c ",c)
	}
	fmt.Printf("\n")
	for i,row := range board{
		fmt.Printf("%d, %v\n",i,row)
		fmt.Printf("    ")
		for j := range row{
			fmt.Printf("%s ",board_letters[j])
		}
		fmt.Printf("\n")
	}
	
	
}