package pieces

import (
	"testing"

	"github.com/stretchr/testify/require"
)


func TestQueen(t *testing.T){
	initialBoard := map[string]string{
		// White pieces
		"a2": "P", "b3": "P", "c3": "P", "d4": "P", "e2": "P", "f2": "P", "g6": "P", "h2": "P",
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
			name:         "Qd1",
			position:     "d1",
			legalSquares: []string{"c1","d2","d3","c2"},
		},
		{
			name:         "Qd8",
			position:     "d8",
			legalSquares: []string{"e7","f6","g5","h4"},
		},
		
		
	}

	PrintBoard(chessBoard)
	for _, squares := range gameState.Board {
		for _, square := range squares {
			if square.Occupied && square.Piece.GetPieceType() == "Q" {
				for _, piecetest := range pieceTests {
					if square.Piece.GetPosition() == piecetest.position {
						t.Run(piecetest.name, func(t *testing.T) {
							require.Equal(t,piecetest.legalSquares,square.Piece.GetLegalSquares(gameState))
						})
					}
				}
			}
		}
	}
}