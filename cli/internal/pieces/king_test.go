package pieces

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKingPiece(t *testing.T){
	kingTests := []struct {
		name         string
		piece        PieceInterface
		legalSquares []string
	}{
		{
			name: "king1",
			piece: &King{
				PieceType: "K",
				Color:     "w",
				Position:  "e1",
			},
			legalSquares: []string{"d1","d2","f1","f2","e2"},
		},
		{
			name: "king2",
			piece: &King{
				PieceType: "K",
				Color:     "w",
				Position:  "e5",
			},
			legalSquares: []string{"e4","d4","d5","d6","e6","f6","f5","f4"},
		},
	}


	for _, tt := range kingTests {
		t.Run(tt.name, func(t *testing.T) {
			gotlegalSquares := tt.piece.GetLegalSquares()

			require.ElementsMatch(t,gotlegalSquares,tt.legalSquares)
			
		})
	}
}