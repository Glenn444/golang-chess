package stockfish

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Stockfish struct {
	stdin  io.WriteCloser
	stdout *bufio.Scanner
}

// NewStockfish starts the Stockfish engine. Returns nil, nil when
// STOCKFISH_ENGINE_PATH is not set (engine disabled, not an error).
func NewStockfish() (*Stockfish, error) {
	stockfishPath := os.Getenv("STOCKFISH_ENGINE_PATH")
	if stockfishPath == "" {
		return nil, nil
	}

	cmd := exec.Command(stockfishPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stockfish stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stockfish stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("stockfish start: %w", err)
	}

	sf := &Stockfish{
		stdin:  stdin,
		stdout: bufio.NewScanner(stdout),
	}

	sf.send("uci")
	sf.waitFor("uciok")
	sf.send("isready")
	sf.waitFor("readyok")
	sf.send("ucinewgame")

	return sf, nil
}

func (sf *Stockfish) send(cmd string) {
	io.WriteString(sf.stdin, cmd+"\n")
}

func (sf *Stockfish) waitFor(expected string) string {
	for sf.stdout.Scan() {
		line := sf.stdout.Text()
		if strings.Contains(line, expected) {
			return line
		}
	}
	return ""
}

// GetBestMove sends the full move history and waits for Stockfish's move.
// Returns an empty string if Stockfish is not available.
func (sf *Stockfish) GetBestMove(moves []string) string {
	if sf == nil {
		return ""
	}

	if len(moves) == 0 {
		sf.send("position startpos")
	} else {
		sf.send("position startpos moves " + strings.Join(moves, " "))
	}

	sf.send("go movetime 1000")

	line := sf.waitFor("bestmove")
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}
