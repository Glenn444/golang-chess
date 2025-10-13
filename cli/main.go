package main

import (
	"fmt"

	"github.com/Glenn444/golang-chess/internal/board"
	"github.com/Glenn444/golang-chess/internal/cli"
)

func main() {
	b := board.Create_board()
	initialBoard_position := board.Initialise_board(b)
	
	//fmt.Printf("%v",initialBoard_position)
	var game = board.GameState{
		CurrentPlayer: "w",
		Board:         initialBoard_position,
	}
	fmt.Printf("Welcome to the chess game cli\n")

	for _,sq := range initialBoard_position{
		for _, s := range sq{
			if s.Occupied{
				fmt.Printf("%v",s.Piece)
			}else{
				fmt.Printf("[] ")
			}
		}
		fmt.Printf("\n")
		
	}
	
	
	cli.Cli(&game)
}
