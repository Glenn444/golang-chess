package cli

import (
	"fmt"
	"os"

	"github.com/Glenn444/golang-chess/internal/board"
)

type CLI struct {
	game     *board.GameState
	commands map[string]Command
}

type Command struct {
	Name        string
	Description string
	Execute     func([]string) error
}

func NewCLI(g *board.GameState) *CLI {
	cli := &CLI{
		game:     g,
		commands: make(map[string]Command),
	}

	cli.registerCommands()
	return cli
}


/*
	Support two commands
	1. exit - close the game loop
	2. pboard - prints current board state
*/
func (c *CLI) registerCommands() {
	c.commands["exit"] = Command{
		Name:        "exit",
		Description: "Exit the Chess game",
		Execute:     c.exitCommand,
	}

	c.commands["pboard"] = Command{
		Name: "pboard",
		Description: "Prints the board current state",
		Execute: c.printBoardState,
	}
}

func (c *CLI) Execute(tokens []string)error{
	cmdName := tokens[0]

	if cmd,exists := c.commands[cmdName];exists{
		return cmd.Execute(tokens)
	}

	//move the piece if it is not a cli command
	board.Move(c.game,cmdName)
	return nil
}

func (c *CLI) exitCommand([] string) error {
	fmt.Printf("Closing the Chess Game... Goodbye!\n")
	os.Exit(0)
	return nil
}

func (c *CLI) printBoardState([] string) error{
	fmt.Printf("      a  b  c  d  e  f  g  h\n")
	for i, row := range c.game.Board {
		
		fmt.Printf("%d", i+1)
		fmt.Printf("    ")
		for _, s := range row {
			
			if s.Occupied{
			fmt.Printf("%v", s.Piece.String())
			}else{
				fmt.Printf("[ ]")
			}
			
		}
	fmt.Printf("\n")

	}
	fmt.Printf("      a  b  c  d  e  f  g  h\n")
	return nil
}



