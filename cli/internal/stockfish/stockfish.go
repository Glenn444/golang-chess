package stockfish

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Stockfish struct{
    stdin io.WriteCloser
    stdout *bufio.Scanner
}

func NewStockfish()*Stockfish{
    stockfishPath := os.Getenv("STOCKFISH_ENGINE_PATH")
    cmd := exec.Command(stockfishPath)

    stdin,err := cmd.StdinPipe()
    if err != nil{
        log.Fatal(err)
    }

    stdout,err1 := cmd.StdoutPipe()
    if err1 != nil{
        log.Fatal(err1)
    }

    if err2 := cmd.Start(); err2 != nil{
        log.Fatal(err2)
    }

    sf := &Stockfish{
        stdin: stdin,
        stdout: bufio.NewScanner(stdout),
    }

    // initialize stockfish
    sf.send("uci")
    sf.waitFor("uciok")

    sf.send("isready")
    sf.waitFor("readyok")

    sf.send("ucinewgame")

    return sf
}

// send a command to stockfish
func (sf *Stockfish) send(cmd string) {
    io.WriteString(sf.stdin, cmd+"\n")
}

// read lines until we see a line containing the expected string
func (sf *Stockfish) waitFor(expected string) string {
    for sf.stdout.Scan() {
        line := sf.stdout.Text()
        fmt.Println("stockfish:", line)
        if strings.Contains(line, expected) {
            return line
        }
    }
    return ""
}

// GetBestMove sends the full move history and waits for stockfish's move
func (sf *Stockfish) GetBestMove(moves []string) string {
   
    if len(moves) == 0 {
        //stockfish starts as white
        sf.send("position startpos")
    } else {
        //stockfish plays black
        sf.send("position startpos moves " + strings.Join(moves, " "))
    }

    // tell stockfish to think for 1 second
    sf.send("go movetime 1000")

    // wait for bestmove response
    line := sf.waitFor("bestmove")

    // line looks like "bestmove d7d5 ponder e2e4"
    parts := strings.Fields(line)
    if len(parts) >= 2 {
        return parts[1] // return just "d7d5"
    }
    return ""
}