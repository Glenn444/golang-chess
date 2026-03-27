package main

import (
	"fmt"
	"log"

	"github.com/Glenn444/golang-chess/internal/board"
	"github.com/Glenn444/golang-chess/internal/cli"
	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil{
		log.Fatalf("error loading .env: %s\n",err)
	}
	
	b := board.Create_board()
	initialBoard_position := board.Initialise_board(b)

	
	fmt.Printf("Welcome to the chess game cli\n")
	

	color,err := board.ChooseColor()
	if err != nil{
		log.Fatalf("%s", err.Error())
	}
	opponent, err2 := board.ChooseGameType()
	if err2 != nil{
		log.Fatalf("%s",err2.Error())
	}

	var game = pieces.GameState{
		CurrentPlayer: "w",
		Board:         initialBoard_position,
		CapturedPieces: make(map[string][]pieces.PieceInterface),
		PlayAgainst: opponent,
		UserColor: color,
	}

	board.PrintBoard(game)
	
	cli.Cli(&game)
}
