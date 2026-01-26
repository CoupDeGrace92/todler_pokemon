package gamelogic

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func GetInput() []string {
	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)
	scanned := scanner.Scan()
	if !scanned {
		return nil
	}
	line := scanner.Text()
	line = strings.TrimSpace(line)
	return strings.Fields(line)
}

func ClientWelcome() {
	fmt.Println("Welcome to P4T - The most complicated pokemon game in the world")
	fmt.Println("...                                                         ...")
	fmt.Println("Here you will have a chance to catch them all, either by memory, or by random!")
	fmt.Println("A list of supported commands:")
	fmt.Println()
	fmt.Println("Name               |Args                    |Usage                             ")
	fmt.Println("_______________________________________________________________________________")
	fmt.Println("catch              |-r:  random mode        |Will catch a pokemon, a random one")
	fmt.Println("                   |<pokemon>:              |if -r is specified, else <pokemon>")
	fmt.Println("quit               |None                    |Will exit p4t                     ")

	fmt.Println()
	fmt.Println()
}
