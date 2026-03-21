package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Glenn444/golang-chess/internal/board"
	//"github.com/Glenn444/golang-chess/internal/cli"
	"github.com/Glenn444/golang-chess/internal/pieces"
)

func main() {
	b := board.Create_board()
	initialBoard_position := board.Initialise_board(b)
	
	//fmt.Printf("%v",initialBoard_position)
	
	fmt.Printf("Welcome to the chess game cli\n")
	

	// for _,sq := range initialBoard_position{
	// 	for _, s := range sq{
	// 		if s.Occupied{
	// 			fmt.Printf("%v",s.Piece)
	// 		}else{
	// 			fmt.Printf("[ ]")
	// 		}
	// 	}
	// 	fmt.Printf("\n")
		
	// }
	var player string
	fmt.Printf("Please select the color you'll be playing w for white,b for black:\n")
	fmt.Printf("play: ")
	scanner := bufio.NewScanner(os.Stdin)
	for{
		if !scanner.Scan(){
			break
		}

		token := cleanInput(scanner.Text())
		if token[0] == "w" || token[0] == "b"{
			player = token[0]
			break
		}else{
			fmt.Printf("token: %v\n",token)
		}
		fmt.Printf("play: ")
	}
	fmt.Printf("player: %s\n",player)


	var game = pieces.GameState{
		CurrentPlayer: player,
		Board:         initialBoard_position,
		CapturedPieces: make(map[string][]pieces.PieceInterface),
	}

	board.PrintBoard(game)
	
	//cli.Cli(&game)
}

func cleanInput(text string)[]string{
	text = (strings.TrimSpace(text))
	//fmt.Printf("%s\n", text)

	return strings.Fields(text) //fields splits on any whitespace
}
