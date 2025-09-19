package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Glenn444/golang-chess/internal/board"
	//"github.com/Glenn444/golang-chess/internal/board"
	//"github.com/Glenn444/golang-chess/utils"
)

type GameState struct {
	CurrentPlayer string
	Board board.Square
}
type cliCommand struct {
	name        string
	description string
	callback    func(cfg GameState,move string) error
}
var c_player = GameState{
		CurrentPlayer: "w",
	}
func main() {

	// initialBoard := board.Create_board()

	// b := board.Initialise_board(initialBoard)
	// for i,row := range b{
	// 	fmt.Printf("%d\n",i)
	// 	fmt.Printf("    ")
	// 	for _,v := range row{
	// 		//positon := fmt.Sprintf("%s%d",board_letters[j],i+1)
	// 		fmt.Printf("%v ",v)

	// 	}
	// 	fmt.Printf("\n")

	// }

	scanner := bufio.NewScanner(os.Stdin)
	supportedCmds := map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
	}
	
	for {
		if c_player.CurrentPlayer == "w" {
			fmt.Printf("White Move > ")
			c_player.CurrentPlayer = "b"
		} else {
			fmt.Printf("Black Move > ")
			c_player.CurrentPlayer = "w"
		}
		if !scanner.Scan() {
			break
		}
		tokens := cleaninput(scanner.Text())
		if len(tokens) == 0 {
			continue
		}
		cmdName := tokens[0]
		if len(tokens) == 1 {
			tokens = append(tokens, " ")
		}

		//fmt.Println("CMD: ",tokens)

		if cmd, ok := supportedCmds[cmdName]; ok {
			

			err := cmd.callback(c_player,tokens[1])
			if err != nil {
				tokens[1] = " "
				fmt.Println("Error:", err)
			}
		} else {
			fmt.Println("Unknown command")
		}
	}
}

func cleaninput(text string) []string {
	
	//implement the logic
	//1. trim space
	//2. to lowercase
	//3. split to whitespace
	text = (strings.TrimSpace(text))
	fmt.Printf("%s\n", text)
	
	return strings.Fields(text) //fields splits on any whitespace
}

func commandExit(_ GameState,_ string) error {
	fmt.Printf("Closing the Chess Game... Goodbye!\n")
	os.Exit(0)
	return nil
}

//pos := utils.Indices_to_chess_notation(0,0)//"a1"
//row_pos,col_pos := utils.Chess_notation_to_indices("a1")
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
