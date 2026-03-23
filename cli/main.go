package main

import (
	"fmt"
	"log"
	

	"github.com/Glenn444/golang-chess/internal/board"
	"github.com/Glenn444/golang-chess/internal/cli"
	"github.com/Glenn444/golang-chess/internal/pieces"
)

func main() {
	b := board.Create_board()
	initialBoard_position := board.Initialise_board(b)
	
	//fmt.Printf("%v",initialBoard_position)
	
	fmt.Printf("Welcome to the chess game cli\n")
	

	fmt.Printf("Please select the color you'll be playing w for white,b for black:\n")
	fmt.Printf("Color to play: ")
	color,err := board.ChooseColor()
	if err != nil{
		log.Fatalf("%s", err.Error())
	}

	var game = pieces.GameState{
		CurrentPlayer: color,
		Board:         initialBoard_position,
		CapturedPieces: make(map[string][]pieces.PieceInterface),
	}

	board.PrintBoard(game)
	
	cli.Cli(&game)
}
