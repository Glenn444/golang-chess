package board

import (
	"testing"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/stretchr/testify/require"
)

func TestCurrentOccupiedPositions(t *testing.T){
	gameState := pieces.GameState{
		CurrentPlayer: "b",
		Board: generateBoardPositions(),
	}

	
	require.Error(t,Move(&gameState,"Nc4"))
	require.NoError(t,Move(&gameState,"Nc6"))
}

func generateBoardPositions()[][]pieces.Square{
	b := Create_board()
	initialBoard := Initialise_board(b)

	initialBoard[4][1] = pieces.Square{
		Occupied: false,
		Piece: nil,
	}
	initialBoard[4][3] = pieces.Square{
		Occupied: true,
		Piece: &pieces.Pawn{
			PieceType: "P",
			Color: "w",
			Position: "e4",
		},
	}

	return initialBoard
}