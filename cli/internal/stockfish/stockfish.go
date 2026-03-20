package cli

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
)

func ExecuteStockfish() {

	cmd := exec.Command("cat")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, "values written to stdin are passed to cmd's standard input")
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", out)
}
func ExecuteStockfishCmd(arg string) {
    cmd := exec.Command("sudo","/Users/mac/Desktop/Learngo/chess_game/stockfish")

    // pipe to write to the program
    stdin, err := cmd.StdinPipe()
    if err != nil {
        log.Fatal(err)
    }

    // pipe to read from the program
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        log.Fatal(err)
    }

    // start the program (non-blocking, unlike CombinedOutput)
    if err := cmd.Start(); err != nil {
        log.Fatal(err)
    }

    // write to stdin
    go func() {
        defer stdin.Close()
        io.WriteString(stdin, arg)
    }()

    // read output line by line as it comes
    scanner := bufio.NewScanner(stdout)
    for scanner.Scan() {
        line := scanner.Text()
        fmt.Println("got:", line)
    }

    // wait for program to finish
    cmd.Wait()
}