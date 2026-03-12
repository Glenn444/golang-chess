package pieces

import (
	"testing"

	"github.com/stretchr/testify/require"
)


func TestKnight(t *testing.T){
	initialBoard := map[string]string{
		// White pieces
		"a2": "P", "b3": "P", "c3": "P", "d2": "P", "e2": "P", "f2": "P", "g2": "P", "h2": "P",
		"a1": "R", "h1": "R",
		"b1": "N", "c4": "N",
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

	chessBoard :=  SetUpBoard(initialBoard)

	gameState := GameState{
		CurrentPlayer: "w",
		Board:        chessBoard,
	}
	pieceTests := []struct {
		name         string
		position     string
		legalSquares []string
	}{
		{
			name:         "Nb1",
			position:     "b1",
			legalSquares: []string{"a3"},
		},
		{
			name:         "Ng1",
			position:     "g1",
			legalSquares: []string{"f6","h6"},
		},
		{
			name: "Nb8",
			position: "b8",
			legalSquares: []string{"a6","c6"},
		},
		{
			name: "Nc4",
			position: "c4",
			legalSquares: []string{"e3","e5","b6","d6","a3","a5","b2"},
		},
		
	}

	PrintBoard(chessBoard)
	for _, squares := range gameState.Board {
		for _, square := range squares {
			if square.Occupied && square.Piece.GetPieceType() == "N" {
				for _, piecetest := range pieceTests {
					if square.Piece.GetPosition() == piecetest.position {
						t.Run(piecetest.name, func(t *testing.T) {
							require.ElementsMatch(t,piecetest.legalSquares,square.Piece.GetLegalSquares(gameState))
						})
					}
				}
			}
		}
	}
}