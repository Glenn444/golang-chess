package pieces

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKingPiece(t *testing.T){
	board := [][]Square{}
	gameState := GameState{
		CurrentPlayer: "w",
		Board:         Initialise_board(board),
	}
	kingTests := []struct {
		name         string
		position     string
		legalSquares []string
	}{
		{
			name:         "bc1",
			position:     "c1",
			legalSquares: []string{""},
		},
		{
			name:         "bf1",
			position:     "f1",
			legalSquares: []string{""},
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

