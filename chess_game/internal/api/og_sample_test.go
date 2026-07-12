package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Glenn444/golang-chess/internal/board"
	"github.com/stretchr/testify/require"
)

// TestOGSampleCards renders one card of each layout to OG_SAMPLE_DIR for
// visual inspection. Skipped unless the env var is set:
//
//	OG_SAMPLE_DIR=/tmp go test ./internal/api -run TestOGSampleCards
func TestOGSampleCards(t *testing.T) {
	dir := os.Getenv("OG_SAMPLE_DIR")
	if dir == "" {
		t.Skip("OG_SAMPLE_DIR not set")
	}
	require.NoError(t, ogFonts())

	cards := map[string]ogCard{
		"live": {
			pillLabel: "LIVE", pillColor: ogRed,
			title:   "glennmakhandia vs deep_thought",
			context: "Move 24 · 10 min game · watch live on chesske.com",
			board:   board.Initialise_board(board.Create_board()),
		},
		"invite": {
			pillLabel: "GAME INVITE", pillColor: ogAmber, pillDark: true,
			title:   "glennmakhandia challenges you",
			context: "10 min game · you play black · chesske.com",
			board:   board.Initialise_board(board.Create_board()),
		},
		"locked": {
			pillLabel: "PRIVATE", pillColor: ogGrey,
			title:   "A private game",
			context: "Only its players can watch · play your own on chesske.com",
		},
	}
	for name, card := range cards {
		png, err := renderOGCard(card)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "og-"+name+".png"), png, 0o644))
	}
}
