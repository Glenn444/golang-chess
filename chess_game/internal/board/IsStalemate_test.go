package board

import (
	"testing"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/stretchr/testify/require"
)

func TestIsStalemate(t *testing.T) {
	tests := []struct {
		name          string
		boardState    map[string]string
		currentPlayer string
		want          bool
	}{
		// ── True stalemates (king not in check, no legal move for any piece) ──
		{
			name:          "Queen + King vs lone King — classic stalemate",
			boardState:    map[string]string{"a8": "k", "b6": "Q", "c7": "K"},
			currentPlayer: "b",
			want:          true,
		},
		{
			name:          "King alone vs queen — all escapes covered",
			boardState:    map[string]string{"a1": "K", "b3": "q"},
			currentPlayer: "w",
			want:          true,
		},
		{
			name:          "King alone vs rook — all escapes covered",
			boardState:    map[string]string{"a1": "K", "b2": "r", "c3": "k"},
			currentPlayer: "w",
			want:          true,
		},

		// ── Not stalemate ────────────────────────────────────────────────────
		{
			name:          "King has legal escape squares",
			boardState:    map[string]string{"a1": "K", "h8": "k"},
			currentPlayer: "w",
			want:          false,
		},
		{
			name:          "King is in check — not stalemate (check, not mate)",
			boardState:    map[string]string{"a1": "K", "h1": "r"},
			currentPlayer: "w",
			want:          false,
		},
		{
			name:          "Checkmate is not stalemate",
			boardState:    map[string]string{"a8": "k", "b7": "Q", "c6": "K"},
			currentPlayer: "b",
			want:          false,
		},
		{
			name:          "Player has other pieces that can move",
			boardState:    map[string]string{"a1": "K", "a2": "P", "h8": "k"},
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

			got := IsStalemate(gameState)
			require.Equal(t, tt.want, got)
		})
	}
}
