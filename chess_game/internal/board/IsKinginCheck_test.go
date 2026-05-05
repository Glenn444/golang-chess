package board

import (
	"testing"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/stretchr/testify/require"
)

func TestIsKinginCheck(t *testing.T) {
	tests := []struct {
		name          string
		boardState    map[string]string
		currentPlayer string
		want          bool
	}{
		// ── King is in check ─────────────────────────────────────────────────
		{
			name:          "Rook gives check along rank",
			boardState:    map[string]string{"a1": "K", "h1": "r"},
			currentPlayer: "w",
			want:          true,
		},
		{
			name:          "Rook gives check along file",
			boardState:    map[string]string{"a1": "K", "a8": "r"},
			currentPlayer: "w",
			want:          true,
		},
		{
			name:          "Bishop gives check on diagonal",
			boardState:    map[string]string{"a1": "K", "h8": "b"},
			currentPlayer: "w",
			want:          true,
		},
		{
			name:          "Queen gives check",
			boardState:    map[string]string{"e1": "K", "e8": "q"},
			currentPlayer: "w",
			want:          true,
		},
		{
			name:          "Knight gives check",
			boardState:    map[string]string{"e1": "K", "f3": "n"},
			currentPlayer: "w",
			want:          true,
		},
		{
			name:          "Pawn gives check to king",
			boardState:    map[string]string{"e4": "K", "d5": "p"},
			currentPlayer: "w",
			want:          true,
		},
		{
			name:          "Black king in check from white bishop",
			boardState:    map[string]string{"e8": "k", "b5": "B"},
			currentPlayer: "b",
			want:          true,
		},
		{
			name:          "Double check — rook and bishop",
			boardState:    map[string]string{"e1": "K", "e8": "r", "a4": "b"},
			currentPlayer: "w",
			want:          true,
		},

		// ── King is NOT in check ─────────────────────────────────────────────
		{
			name:          "No enemy pieces on board",
			boardState:    map[string]string{"e1": "K", "e8": "k"},
			currentPlayer: "w",
			want:          false,
		},
		{
			name:          "Enemy piece nearby but not attacking",
			boardState:    map[string]string{"e4": "K", "c6": "n"},
			currentPlayer: "w",
			want:          false,
		},
		{
			name:          "Rook blocked by own piece — no check",
			boardState:    map[string]string{"a1": "K", "a4": "R", "a8": "r"},
			currentPlayer: "w",
			want:          false,
		},
		{
			name:          "Bishop blocked — no check",
			boardState:    map[string]string{"a1": "K", "d4": "P", "h8": "b"},
			currentPlayer: "w",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board := pieces.SetUpBoard(tt.boardState)
			gameState := pieces.GameState{
				CurrentPlayer: tt.currentPlayer,
				Board:         board,
			}

			got := IsKinginCheck(gameState)
			require.Equal(t, tt.want, got)
		})
	}
}
