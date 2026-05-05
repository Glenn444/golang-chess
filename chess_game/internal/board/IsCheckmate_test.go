package board

import (
	"testing"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/stretchr/testify/require"
)

func TestIsCheckMate(t *testing.T) {
	checkMateTests := []struct {
		name          string
		boardState    map[string]string
		currentPlayer string
		want          bool
	}{
		// ── True checkmates ──────────────────────────────────────────────────
		{
			name:          "Back-rank mate: two rooks vs lone king",
			boardState:    map[string]string{"a1": "K", "h1": "r", "h2": "r"},
			currentPlayer: "w",
			want:          true,
		},
		{
			name:          "Queen + King vs lone King (corner)",
			boardState:    map[string]string{"a8": "k", "b7": "Q", "c6": "K"},
			currentPlayer: "b",
			want:          true,
		},
		{
			name:          "Smothered mate — own pieces block escape",
			boardState:    map[string]string{"h8": "k", "g8": "r", "g7": "p", "h7": "p", "f7": "N", "e1": "K"},
			currentPlayer: "b",
			want:          true,
		},
		{
			name:          "Back-rank mate with blocked pawns",
			boardState:    map[string]string{"e8": "k", "d7": "p", "e7": "p", "f7": "p", "h8": "R", "e1": "K"},
			currentPlayer: "b",
			want:          true,
		},
		{
			name:          "Anastasia's mate: rook + knight",
			boardState:    map[string]string{"h8": "k", "g7": "p", "h5": "R", "e7": "N", "e1": "K"},
			currentPlayer: "b",
			want:          true,
		},

		// ── Not checkmate (can escape / capture / block) ─────────────────────
		{
			name:          "King in check but can escape",
			boardState:    map[string]string{"a1": "K", "h1": "r"},
			currentPlayer: "w",
			want:          false,
		},
		{
			name:          "King in check but can capture attacker",
			boardState:    map[string]string{"a1": "K", "a2": "r"},
			currentPlayer: "w",
			want:          false,
		},
		{
			name:          "Friendly piece can block the check",
			boardState:    map[string]string{"a1": "K", "h1": "r", "d4": "R"},
			currentPlayer: "w",
			want:          false,
		},
		{
			name:          "King not in check at all — normal position",
			boardState:    map[string]string{"e1": "K", "e8": "k"},
			currentPlayer: "w",
			want:          false,
		},
		{
			name:          "Stalemate is not checkmate",
			boardState:    map[string]string{"a8": "k", "b6": "Q", "c7": "K"},
			currentPlayer: "b",
			want:          false,
		},
	}

	for _, tt := range checkMateTests {
		t.Run(tt.name, func(t *testing.T) {
			board := pieces.SetUpBoard(tt.boardState)
			gameState := pieces.GameState{
				CurrentPlayer: tt.currentPlayer,
				Board:         board,
			}

			got := IsCheckmate(gameState)
			require.Equal(t, tt.want, got)
		})
	}
}
