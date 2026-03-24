package cli

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Glenn444/golang-chess/internal/board"
	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/internal/stockfish"
)


func Cli(g *pieces.GameState) {
	scanner := bufio.NewScanner(os.Stdin)

	cliApp := NewCLI(g)
	for {
		if g.CurrentPlayer == "b" && len(g.StockfishGame) == 0 {
			fmt.Printf("stockfish to play\n")
			sf := stockfish.NewStockfish()
			stockfishMove := sf.GetBestMove(g.StockfishGame)
			_, err := board.Move(g, stockfishMove)
			if err != nil {
				log.Fatalf("%s",err)
			}
			g.StockfishGame = append(g.StockfishGame, stockfishMove)
		}else if g.CurrentPlayer == "w" {
			fmt.Printf("White Move > ")
		} else {
			fmt.Printf("Black Move > ")
		}
		if !scanner.Scan() {
			break
		}

		tokens := cleanInput(scanner.Text())
		if len(tokens) == 0 {
			continue
		}

		
		if err := cliApp.Execute(tokens); err != nil{
			fmt.Printf("Error: %v\n",err)
		}

	}
}

func cleanInput(text string) []string {

	//implement the logic
	//1. trim space
	//2. to lowercase
	//3. split to whitespace
	text = (strings.TrimSpace(text))
	//fmt.Printf("%s\n", text)

	return strings.Fields(text) //fields splits on any whitespace
}

