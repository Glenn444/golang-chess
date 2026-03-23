package board

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

func ChooseColor()(string,error){
	scanner := bufio.NewScanner(os.Stdin)
	for{
		if !scanner.Scan(){
			break
		}

		token := cleanInput(scanner.Text())
		if token[0] == "w" || token[0] == "b"{
			return token[0],nil
		}
	}
	return "",errors.New("selected no color w or b")

}

func cleanInput(text string)[]string{
	text = (strings.TrimSpace(text))
	//fmt.Printf("%s\n", text)

	return strings.Fields(text) //fields splits on any whitespace
}