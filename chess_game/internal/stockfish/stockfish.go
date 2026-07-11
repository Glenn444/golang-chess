package stockfish

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	// How long to wait for engine responses (uciok/readyok/bestmove). The
	// search itself is bounded by "go movetime", so this only trips when the
	// engine is dead or wedged.
	responseTimeout = 10 * time.Second
	moveTimeMs      = 1000
)

// uciMove matches coordinate moves like e2e4 and promotions like e7e8q —
// history entries are joined into a UCI command, so anything else (spaces,
// newlines) could inject arbitrary engine commands.
var uciMove = regexp.MustCompile(`^[a-h][1-8][a-h][1-8][qrbn]?$`)

type Stockfish struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
	lines chan string
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
		cmd:   cmd,
		stdin: stdin,
		lines: make(chan string, 64),
	}

	// Pump engine output into a channel so reads can time out instead of
	// blocking forever on a dead engine.
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			sf.lines <- scanner.Text()
		}
		close(sf.lines)
	}()

	for _, step := range []struct{ cmd, expect string }{
		{"uci", "uciok"},
		{"isready", "readyok"},
	} {
		if err := sf.send(step.cmd); err != nil {
			sf.Close()
			return nil, err
		}
		if _, err := sf.waitFor(step.expect); err != nil {
			sf.Close()
			return nil, err
		}
	}
	if err := sf.send("ucinewgame"); err != nil {
		sf.Close()
		return nil, err
	}

	return sf, nil
}

// Close asks the engine to quit and reaps the process so it never leaks or
// zombies. Safe to call on a nil receiver.
func (sf *Stockfish) Close() {
	if sf == nil || sf.cmd == nil {
		return
	}
	_ = sf.send("quit")
	sf.stdin.Close()

	done := make(chan struct{})
	go func() {
		sf.cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		sf.cmd.Process.Kill()
		<-done
	}
	sf.cmd = nil
}

func (sf *Stockfish) send(cmd string) error {
	_, err := io.WriteString(sf.stdin, cmd+"\n")
	if err != nil {
		return fmt.Errorf("stockfish write %q: %w", cmd, err)
	}
	return nil
}

// waitFor reads engine output until a line containing expected appears, the
// engine exits, or the response timeout elapses.
func (sf *Stockfish) waitFor(expected string) (string, error) {
	deadline := time.After(responseTimeout)
	for {
		select {
		case line, ok := <-sf.lines:
			if !ok {
				return "", errors.New("stockfish: engine exited unexpectedly")
			}
			if strings.Contains(line, expected) {
				return line, nil
			}
		case <-deadline:
			return "", fmt.Errorf("stockfish: timed out waiting for %q", expected)
		}
	}
}

// SetSkillLevel sets the UCI "Skill Level" option (0 weakest — 20 strongest)
// and waits for the engine to acknowledge readiness.
func (sf *Stockfish) SetSkillLevel(level int) error {
	if sf == nil {
		return nil
	}
	if level < 0 {
		level = 0
	}
	if level > 20 {
		level = 20
	}
	if err := sf.send(fmt.Sprintf("setoption name Skill Level value %d", level)); err != nil {
		return err
	}
	if err := sf.send("isready"); err != nil {
		return err
	}
	_, err := sf.waitFor("readyok")
	return err
}

// GetBestMove sends the full move history and waits for Stockfish's move.
// Returns an empty string if Stockfish is not available.
func (sf *Stockfish) GetBestMove(moves []string) (string, error) {
	if sf == nil {
		return "", nil
	}

	for _, m := range moves {
		if !uciMove.MatchString(m) {
			return "", fmt.Errorf("stockfish: invalid move in history: %q", m)
		}
	}

	if len(moves) == 0 {
		if err := sf.send("position startpos"); err != nil {
			return "", err
		}
	} else {
		if err := sf.send("position startpos moves " + strings.Join(moves, " ")); err != nil {
			return "", err
		}
	}

	if err := sf.send(fmt.Sprintf("go movetime %d", moveTimeMs)); err != nil {
		return "", err
	}

	line, err := sf.waitFor("bestmove")
	if err != nil {
		return "", err
	}
	parts := strings.Fields(line)
	if len(parts) >= 2 && parts[1] != "(none)" {
		return parts[1], nil
	}
	return "", nil
}
