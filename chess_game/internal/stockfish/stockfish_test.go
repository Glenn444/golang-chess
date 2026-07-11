package stockfish

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// TestEngineSmoke exercises the real engine end-to-end: spawn, set skill
// level, get an opening move, get a reply after 1.e4, and shut down cleanly.
// Skipped when no engine binary is available.
func TestEngineSmoke(t *testing.T) {
	if os.Getenv("STOCKFISH_ENGINE_PATH") == "" {
		local, err := filepath.Abs("stockfish-engine")
		if err != nil {
			t.Skip("no engine binary available")
		}
		if _, err := os.Stat(local); err != nil {
			t.Skip("no engine binary available (set STOCKFISH_ENGINE_PATH)")
		}
		t.Setenv("STOCKFISH_ENGINE_PATH", local)
	}

	sf, err := NewStockfish()
	if err != nil {
		t.Fatalf("NewStockfish: %v", err)
	}
	if sf == nil {
		t.Skip("engine disabled")
	}
	defer sf.Close()

	if err := sf.SetSkillLevel(3); err != nil {
		t.Fatalf("SetSkillLevel: %v", err)
	}

	uci := regexp.MustCompile(`^[a-h][1-8][a-h][1-8][qrbn]?$`)

	move, err := sf.GetBestMove(nil)
	if err != nil {
		t.Fatalf("GetBestMove(startpos): %v", err)
	}
	if !uci.MatchString(move) {
		t.Fatalf("expected a UCI move, got %q", move)
	}

	reply, err := sf.GetBestMove([]string{"e2e4"})
	if err != nil {
		t.Fatalf("GetBestMove(after e4): %v", err)
	}
	if !uci.MatchString(reply) {
		t.Fatalf("expected a UCI reply, got %q", reply)
	}

	// History validation must reject injection attempts.
	if _, err := sf.GetBestMove([]string{"e2e4\nquit"}); err == nil {
		t.Error("malformed history must be rejected")
	}
}
