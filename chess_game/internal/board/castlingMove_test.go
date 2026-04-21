package board

import (
	"testing"

	"github.com/Glenn444/golang-chess/internal/pieces"
)

func TestCastling(t *testing.T) {

	castleTests := []struct{
		name string
		boardState map[string]string
		currentPlayer string
		move string
		expectedValid bool
		castling pieces.Castling
		

	}{
		{
			name: "King Side Castling - path clear",
			boardState: map[string]string{"e1": "K", "h1": "R"},
			expectedValid: true,
			move: "O-O",
			currentPlayer: "w",
			castling:      pieces.Castling{},
		},
		{
			name: "White queenside castling - path clear",
			boardState: map[string]string{"e1": "K", "a1": "R"},
			move: "O-O-O",
			currentPlayer: "w",
			expectedValid: true,
			castling:      pieces.Castling{},
		},
		{
			name:          "Black kingside castling — path clear",
			boardState:    map[string]string{"e8": "k", "h8": "r"},
			currentPlayer: "b",
			move: "O-O",
			expectedValid: true,
			castling:      pieces.Castling{},
		},
		{
			name:          "Black queenside castling — path clear",
			boardState:    map[string]string{"e8": "k", "a8": "r"},
			currentPlayer: "b",
				move: "O-O-O",
			expectedValid: true,
			castling:      pieces.Castling{},
		},
		{
			name:          "Both sides can castle",
			boardState:    map[string]string{"e1": "K", "a1": "R", "h1": "R", "e8": "k", "a8": "r", "h8": "r"},
			currentPlayer: "w",
			move: "O-O",
			expectedValid: true,
			castling:      pieces.Castling{},
		},

		// ❌ Blocked paths
		{
			name:          "White kingside blocked by knight on g1",
			boardState:    map[string]string{"e1": "K", "h1": "R", "g1": "N"},
			currentPlayer: "w",
			expectedValid: false,
			move: "O-O",
			castling:      pieces.Castling{},
		},
		{
			name:          "White queenside blocked by bishop on c1",
			boardState:    map[string]string{"e1": "K", "a1": "R", "c1": "B"},
			currentPlayer: "w",
			expectedValid: false,
				move: "O-O-O",
			castling:      pieces.Castling{},
		},

		// ❌ Check conditions
		{
			name:          "King in check — castling not allowed",
			boardState:    map[string]string{"e1": "K", "h1": "R", "e8": "r"},
			currentPlayer: "w",
				move: "O-O",
			expectedValid: false,
		},
		{
			name:          "King passes through attacked square f1",
			boardState:    map[string]string{"e1": "K", "h1": "R", "f8": "r"},
			currentPlayer: "w",
				move: "O-O",
			expectedValid: false,
		},
		{
			name:          "King lands on attacked square g1",
			boardState:    map[string]string{"e1": "K", "h1": "R", "g8": "r"},
			currentPlayer: "w",
				move: "O-O",
			expectedValid: false,
			castling:      pieces.Castling{},
		},

		// ❌ Castling rights revoked
		{
			name:          "White kingside right revoked — king moved previously",
			boardState:    map[string]string{"e1": "K", "h1": "R"},
			currentPlayer: "w",
				move: "O-O",
			expectedValid: false,
			castling:      pieces.Castling{WhiteKingMoved: true},
		},
		{
			name:          "White queenside right revoked — rook moved previously",
			boardState:    map[string]string{"e1": "K", "a1": "R"},
			currentPlayer: "w",
				move: "O-O",
			expectedValid: false,
			castling:      pieces.Castling{WhiteRookQueensideMoved: true},
		},
	}

	for _,tt := range castleTests{
		t.Run(tt.name,func(t *testing.T) {
			boardState := pieces.SetUpBoard(tt.boardState)
			gameState := pieces.GameState{
				CurrentPlayer: tt.currentPlayer,
				Board: boardState,
				Castle: tt.castling,
			}
			got := true

			err := CastlingMove(&gameState,tt.move)
			if err != nil{
				got = false
			}

			if got != tt.expectedValid{
				t.Errorf("expected %v,got %v (err: %v)",tt.expectedValid,got,err)
			}

		})
	}
	

}
