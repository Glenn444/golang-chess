package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Glenn444/golang-chess/internal/board"
)

type cliCommand struct {
	name        string
	description string
	callback    func(cfg board.GameState, move string) error
}

func Cli(g *board.GameState) {
	scanner := bufio.NewScanner(os.Stdin)
	supportedCmds := map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Game",
			callback:    commandExit,
		},
	}

	for {
		if g.CurrentPlayer == "w" {
			fmt.Printf("White Move > ")
			g.CurrentPlayer = "b"
		} else {
			fmt.Printf("Black Move > ")
			g.CurrentPlayer = "w"
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

		if cmd, ok := supportedCmds[cmdName]; ok {

			err := cmd.callback(*g, tokens[1])
			if err != nil {
				tokens[1] = " "
				fmt.Println("Error:", err)
			}
		}else{
			board.Move(g,cmdName)
		}
	}
}

func cleaninput(text string) []string {

	//implement the logic
	//1. trim space
	//2. to lowercase
	//3. split to whitespace
	text = (strings.TrimSpace(text))
	//fmt.Printf("%s\n", text)

	return strings.Fields(text) //fields splits on any whitespace
}

func commandExit(_ board.GameState, _ string) error {
	fmt.Printf("Closing the Chess Game... Goodbye!\n")
	os.Exit(0)
	return nil
}


/*
	Support two commands
	1. exit - close the game loop
	2. pboard - print current board state
*/