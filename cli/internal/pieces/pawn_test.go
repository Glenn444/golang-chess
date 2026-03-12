package pieces

import (
	"testing"

	"github.com/stretchr/testify/require"
)


func TestPawn(t *testing.T){
	initialBoard := map[string]string{
		// White pieces
		"a2": "P", "b3": "P", "c3": "P", "d4": "P", "e2": "P", "f2": "P", "g2": "P", "h2": "P",
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
			name:         "Pc3",
			position:     "c3",
			legalSquares: []string(nil),
		},
		{
			name:         "Pd4",
			position:     "d4",
			legalSquares: []string{"d5","e5"},
		},
		{
			name: "Pc7",
			position: "c7",
			legalSquares: []string{"c6","c5"},
		},
		{
			name: "Pb3",
			position: "b3",
			legalSquares: []string{"b4",},
		},
		{
			name:"Ph7",
			position: "h7",
			legalSquares: []string{"h6","h5"},
		},
		
	}

	PrintBoard(chessBoard)
	for _, squares := range gameState.Board {
		for _, square := range squares {
			if square.Occupied && square.Piece.GetPieceType() == "P" {
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