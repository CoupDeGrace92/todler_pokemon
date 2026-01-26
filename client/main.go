package main

import (
	gl "coupdegrace92/pokemon_for_todlers/gamelogic"
	"fmt"
)

func main() {
	fmt.Println("Welcome to P4T!")
	fmt.Println("Please enter your username: ")
	user := gl.GetInput()
	if len(user) != 1 {
		fmt.Println("You must enter a username")
		return
	}
	fmt.Println("Please enter your password: ")

	pass := gl.GetInput()
	if len(pass) != 1 {
		fmt.Println("You must enter your password")
		return
	}
	for i := 0; i < 4; i++ {
		fmt.Printf("\033[1A\033[K") //ASCII escape codes to go up a line and delete it
	}
	gl.ClientWelcome()
	//LOGIN HANDLER HERE

repl: //This is so we can break the outerloop insteaad of the switch statement
	for {
		cmd := gl.GetInput()
		switch cmd[0] {
		case "quit":
			fmt.Println("Exiting P4T...")
			break repl
		default:
			fmt.Println("Error: Unrecognized Command")
		}
	}
	//Update DB before exitting program

}
