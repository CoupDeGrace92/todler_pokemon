package main

import (
	gl "coupdegrace92/pokemon_for_todlers/gamelogic"
	"fmt"

	"log"
	"strconv"

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
		for i := 0; i < 7; i++ {
			fmt.Printf("\033[1A\033[K") //ASCII escape codes to go up a line and delete it
		}
		if len(u[0]) == 0 || len(p[0]) == 0 {
			fmt.Println("Must enter a username and password")
		}
		err, s := gl.Register(u[0], p[0])
		if err != nil {
			log.Println(err)
			return
		}
		if s != true {
			fmt.Println("Failed to register: Goodbye!")
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
	err, s := gl.Login(user[0], pass[0])
	if err != nil {
		fmt.Println("Error logging in: ", err)
	}
	if s != true {
		fmt.Println("Failed to login - Goodbye!")
		return
	}

	gl.ClientWelcome()

	//Here we are going to get the states for the user:
	pokedex := make(map[string]int)
	userPokes, err := gl.GetTeam()
	if err != nil {
		log.Println("Error getting users pokemon: ", err)
	}
	for _, poke := range userPokes {
		gl.AddToMap(poke.Name, pokedex)
	}

	newPoke := make(map[string]int)

repl: //This is so we can break the outerloop insteaad of the switch statement
	for {
		cmd := gl.GetInput()
		switch cmd[0] {
		case "quit":
			fmt.Println("Saving pokedex...")
			err = gl.UpdateTeam(newPoke)
			if err != nil {
				fmt.Println("Error saving pokedex: ", err)
			}
			fmt.Println("Exiting P4T...")
			break repl
		case "catch":
			pokeID := 0
			if cmd[1] == "-r" {
				pokeID = gl.CatchRandom()
				fmt.Println("CATCHING POKEMON ", pokeID)
				strID := strconv.Itoa(pokeID)
				poke, err := gl.GetPokemon(strID)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println("CAUGHT: ", poke.Name)
				gl.AddToMap(poke.Name, pokedex)
				gl.AddToMap(poke.Name, newPoke)
				continue
			}
			poke, err := gl.GetPokemon(cmd[1])
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("CAUGHT: ", poke.Name)
			gl.AddToMap(poke.Name, pokedex)
			gl.AddToMap(poke.Name, newPoke)
		case "reset":
			fmt.Println("About to reset user team, are you sure? [y/n]")
			confirm := gl.GetInput()
			if confirm[0] != "y" && confirm[0] != "yes" {
				continue
			}
			fmt.Println("Okay - deleting users pokedex - goodluck on your new journey!")
			gl.Reset(user[0])
			pokedex = make(map[string]int)
			newPoke = make(map[string]int)
		case "pokedex":
			gl.DisplayPokedex(pokedex)
		default:
			fmt.Println("Error: Unrecognized Command")
		}
	}
	//Update DB before exitting program

}
