package main

import (
	gl "coupdegrace92/pokemon_for_todlers/gamelogic"
	"fmt"

	"log"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("client/.env")
	fmt.Println("Welcome to P4T!")
	fmt.Println("Are you a new user or a returning user:")
	fmt.Println("Type: returning for returning user, type: new for new user")
	newReturn := gl.GetInput()
	if newReturn[0] == "new" {
		fmt.Println("Please enter desired username:")
		u := gl.GetInput()
		fmt.Println("Please enter desired password")
		p := gl.GetInput()
		for i := 0; i < 4; i++ {
			fmt.Printf("\033[1A\033[K") //ASCII escape codes to go up a line and delete it
		}
		if len(u[0]) == 0 || len(p[0]) == 0 {
			fmt.Println("Must enter a username and password")
		}
		err := gl.Register(u[0], p[0])
		if err != nil {
			log.Println(err)
			return
		}
	}
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
	err := gl.Login(user[0], pass[0])
	if err != nil {
		fmt.Println("Error logging in: ", err)
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
		case "catch":
			pokeID := 0
			if cmd[1] == "-r" {
				pokeID = gl.CatchRandom()
				fmt.Println("CATCHING POKEMON ", pokeID)
			}
		default:
			fmt.Println("Error: Unrecognized Command")
		}
	}
	//Update DB before exitting program

}
