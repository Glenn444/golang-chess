package pieces

import (
	"testing"

	"github.com/stretchr/testify/require"
)



func TestKingPiece(t *testing.T){
	initialBoard := map[string]string{
		// White pieces
		"a2": "P", "b3": "P", "c3": "P", "d2": "P", "e2": "P", "f2": "P", "g2": "P", "h2": "P",
		"a1": "R", "h1": "R",
		"b1": "N", "g1": "N",
		"e4": "B", "f1": "B",
		"d1": "Q",
		"c5": "K",

		// Black pieces
		"a7": "p", "b7": "p", "c7": "p", "d7": "p", "e5": "p", "f7": "p", "g7": "p", "h7": "p",
		"a8": "r", "h8": "r",
		"b8": "n", "g8": "n",
		"c8": "b", "f8": "b",
		"d8": "q",
		"e8": "k",
	}

	chessBoard :=  SetUpBoard(initialBoard)

	gameState := GameState{
		CurrentPlayer: "w",
		Board:        chessBoard,
	}
	kingTests := []struct {
		name         string
		position     string
		legalSquares []string
	}{
		{
			name:         "Kc5",
			position:     "c5",
			legalSquares: []string{"d5", "b5", "c6", "c4", "d6", "d4", "b6", "b4"},
		},
		{
			name:         "Ke8",
			position:     "e8",
			legalSquares: []string{"e7"},
		},
		
	}

	PrintBoard(chessBoard)
	for _, squares := range gameState.Board {
		for _, square := range squares {
			if square.Occupied && square.Piece.GetPieceType() == "K" {
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

