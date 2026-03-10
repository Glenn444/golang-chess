package pieces

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBishop(t *testing.T) {
	board := Create_board()
	gameState := GameState{
		CurrentPlayer: "w",
		Board:         Initialise_board(board),
	}
	bishopTests := []struct {
		name         string
		position     string
		legalSquares []string
	}{
		{
			name:         "bb2",
			position:     "b2",
			legalSquares: []string(nil),
		},
		{
			name:         "bf1",
			position:     "f1",
			legalSquares: []string(nil),
		},
		{
			name:         "bc8",
			position:     "c8",
			legalSquares: []string(nil),
		},
		{
			name:         "bf8",
			position:     "f8",
			legalSquares: []string(nil),
		},
	}

	PrintBoard()
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
