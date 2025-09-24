package main

import (
	"github.com/Glenn444/golang-chess/internal/board"
	"github.com/Glenn444/golang-chess/internal/cli"
	
)

func main() {
	b := board.Create_board()
	initialBoard_position := board.Initialise_board(b)
	var game = board.GameState{
		CurrentPlayer: "w",
		Board:         initialBoard_position,
	}

	cli.Cli(&game)
}
