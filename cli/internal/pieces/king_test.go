package pieces

import (
	"testing"

	"github.com/stretchr/testify/require"
)



func TestKingPiece(t *testing.T){
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

	gameState := GameState{
		CurrentPlayer: "w",
		Board:         SetUpBoard(initialBoard),
	}
	kingTests := []struct {
		name         string
		position     string
		legalSquares []string
	}{
		{
			name:         "Ke1",
			position:     "e1",
			legalSquares: []string(nil),
		},
		{
			name:         "Ke8",
			position:     "e8",
			legalSquares: []string(nil),
		},
		
	}

	for _, squares := range gameState.Board {
		for _, square := range squares {
			if square.Piece.GetPieceType() == "K" {
				for _, ktest := range kingTests {
					if square.Piece.GetPosition() == ktest.position {
						t.Run(ktest.name, func(t *testing.T) {
							require.Equal(t,ktest.legalSquares,square.Piece.GetLegalSquares(gameState))
						})
					}
				}
			}
		}
	}
}

