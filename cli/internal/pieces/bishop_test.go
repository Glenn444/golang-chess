package pieces

import (
	"testing"

	"github.com/stretchr/testify/require"
)



func TestBishop(t *testing.T) {
	initialBoard := map[string]string{
		// White pieces
		"a2": "P", "b3": "P", "c6": "P", "d2": "P", "e2": "P", "f2": "P", "g2": "P", "h2": "P",
		"a1": "R", "h1": "R",
		"b1": "N", "g1": "N",
		"e4": "B", "f1": "B",
		"d1": "Q",
		"e1": "K",

		// Black pieces
		"a7": "p", "b7": "p", "c7": "p", "d7": "p", "e5": "p", "f7": "p", "g7": "p", "h7": "p",
		"a8": "r", "h8": "r",
		"b8": "n", "g8": "n",
		"c8": "b", "f8": "b",
		"d8": "q",
		"e8": "k",
	}
	

	boardState := SetUpBoard(initialBoard)
	gameState := GameState{
		CurrentPlayer: "w",
		Board:        boardState ,
	}
	bishopTests := []struct {
		name         string
		position     string
		legalSquares []string
	}{
		{
			name:         "Be4",
			position:     "e4",
			legalSquares: []string{"f5", "g6", "h7", "d5", "f3", "d3", "c2"},
		},
		{
			name:         "Bf1",
			position:     "f1",
			legalSquares: []string(nil),
		},
		{
			name:         "Bc8",
			position:     "c8",
			legalSquares: []string(nil),
		},
		{
			name:         "Bf8",
			position:     "f8",
			legalSquares: []string{"e7", "d6", "c5", "b4", "a3"},
		},
	}

	//PrintBoard(boardState)
	for _, squares := range gameState.Board {
		for _, square := range squares {
			if square.Occupied && square.Piece.GetPieceType() == "B" {
				for _, btest := range bishopTests {
					if square.Piece.GetPosition() == btest.position {
						t.Run(btest.name, func(t *testing.T) {
							//fmt.Printf("bishop legalsquares: %v\n",square.Piece.GetLegalSquares(gameState))
							require.Equal(t,btest.legalSquares,square.Piece.GetLegalSquares(gameState))
						})
					}
				}
			}
		}
	}
}


