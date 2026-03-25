package cli

import (
	"fmt"
	"os"

	"github.com/Glenn444/golang-chess/internal/board"
	"github.com/Glenn444/golang-chess/internal/pieces"
)

type CLI struct {
	game     *pieces.GameState
	commands map[string]Command
}

type Command struct {
	Name        string
	Description string
	Execute     func([]string) error
}

func NewCLI(g *pieces.GameState) *CLI {
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
		Name:        "pboard",
		Description: "Prints the board current state",
		Execute:     c.printBoardState,
	}
}

func (c *CLI) Execute(tokens []string) error {
	cmdName := tokens[0]

	if cmd, exists := c.commands[cmdName]; exists {
		return cmd.Execute(tokens)
	}

	//valid chess move
	move := cmdName

	//move the piece if it is not a cli command
	switch c.game.PlayAgainst {
	case "person":
		err := board.Move(c.game, move)
		c.printBoardState(nil)
		return err
	case "stockfish":
		//user playing
		if  c.game.UserColor == c.game.CurrentPlayer {
			fmt.Printf("move: %s\n",move)
			err := board.Move(c.game, move)
			return err
		} 
	}
	return nil
}

func (c *CLI) exitCommand([]string) error {
	fmt.Printf("Closing the Chess Game... Goodbye!\n")
	os.Exit(0)
	return nil
}

func (c *CLI) printBoardState([]string) error {
	var sumB int64
	var sumW int64

	fmt.Printf("Game Points w vs b\n")
	for player, capturedPieces := range c.game.CapturedPieces {

		if player == "w" {
			for _, piece := range capturedPieces {
				sumW = sumW + piece.GetPiecePoints()
			}
		} else {
			for _, piece := range capturedPieces {
				sumB = sumB + piece.GetPiecePoints()
			}
		}
	}
	fmt.Printf("White Points: %d, \t Black Points: %d\n", sumW, sumB)
	if board.IsKinginCheck(*c.game) {
		fmt.Printf("Check!!!!!!!!!!!\n")
	}
	board.PrintBoard(*c.game)
	return nil
}
