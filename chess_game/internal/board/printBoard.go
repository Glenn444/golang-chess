package board

import (
	"fmt"

	"github.com/Glenn444/golang-chess/internal/pieces"
)

func PrintBoard(game pieces.GameState) {
	fmt.Printf("      a  b  c  d  e  f  g  h\n")
	if game.CurrentPlayer == "b" {
		for i, row := range game.Board {

			fmt.Printf("%d", i+1)
			fmt.Printf("    ")
			for _, s := range row {

				if s.Occupied {
					fmt.Printf("%v", s.Piece.String())
				} else {
					fmt.Printf("[ ]")
				}

			}
			fmt.Printf("\n")

		}

	}else{
		for i:=7;i >= 0;i--{
			fmt.Printf("%d",i+1)
				fmt.Printf("    ")
			for j:=0;j < 8; j ++{
				
				if game.Board[i][j].Occupied{
					fmt.Printf("%v",game.Board[i][j].Piece.String())
				}else{
					fmt.Printf("[ ]")
				}
			}
			fmt.Printf("\n")
		}

	}

		fmt.Printf("      a  b  c  d  e  f  g  h\n")

}
