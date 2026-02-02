package main

import (
	gl "coupdegrace92/pokemon_for_todlers/gamelogic"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"log"
	"strconv"
)

type Game struct{}

func (g *Game) Update() errpr {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	g := &Game{}
	cfg := gl.InitializeClientEnv()
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
	err, s := cfg.Login(user[0], pass[0])
	if err != nil {
		fmt.Println("Error logging in: ", err)
	}
	if s != true {
		fmt.Println("Failed to login - Goodbye!")
		return
	}

	gl.ClientWelcome()

	//Here we are going to get the states for the user:
	pokedex := make(map[string]*gl.PokemonPlusCount)
	userPokes, err := cfg.GetTeam()
	if err != nil {
		log.Println("Error getting users pokemon: ", err)
	}
	for _, poke := range userPokes {
		p := gl.RecievedPokemonToPokemon(poke)
		gl.AddToMap(p, pokedex)
	}

	newPoke := make(map[string]int)

repl: //This is so we can break the outerloop insteaad of the switch statement
	for {
		cmd := gl.GetInput()
		switch cmd[0] {
		case "quit":
			fmt.Println("Saving pokedex...")
			err = cfg.UpdateTeam(newPoke)
			if err != nil {
				fmt.Println("Error saving pokedex: ", err)
			}
			fmt.Println("Exiting P4T...")
			break repl
		case "catch":
			pokeID := 0
			if cmd[1] == "-r" {
				pokeID = gl.CatchRandom()
				strID := strconv.Itoa(pokeID)
				poke, err := cfg.GetPokemon(strID)
				if err != nil {
					fmt.Println(err)
					continue
				}
				s, e := gl.PokeToTypeColor(poke)
				text := gl.StringGradient(poke.Name, s, e)
				fmt.Println("CAUGHT: ", text)
				gl.AddToMap(poke, pokedex)
				gl.AddToStringMap(poke, newPoke)
				continue
			}
			poke, err := cfg.GetPokemon(cmd[1])
			if err != nil {
				fmt.Println(err)
				continue
			}
			s, e := gl.PokeToTypeColor(poke)
			text := gl.StringGradient(poke.Name, s, e)
			fmt.Println("CAUGHT: ", text)
			gl.AddToMap(poke, pokedex)
			gl.AddToStringMap(poke, newPoke)
		case "reset":
			fmt.Println("About to reset user team, are you sure? [y/n]")
			confirm := gl.GetInput()
			if confirm[0] != "y" && confirm[0] != "yes" {
				continue
			}
			fmt.Println("Okay - deleting users pokedex - goodluck on your new journey!")
			cfg.Reset(user[0])
			pokedex = make(map[string]*gl.PokemonPlusCount)
			newPoke = make(map[string]int)
		case "pokedex":
			gl.DisplayPokedex(pokedex)
		default:
			fmt.Println("Error: Unrecognized Command")
		}
	}
	//Update DB before exitting program

}
