package main

import (
	"coupdegrace92/pokemon_for_todlers/client/gui"
	gl "coupdegrace92/pokemon_for_todlers/gamelogic"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"log"
	"strconv"
)

func main() {
	cfg := gl.InitializeClientEnv()
	core := &gui.Core{}
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
	core.PokeDex = pokedex
	core.NewPokemon = newPoke

repl: //This is so we can break the outerloop insteaad of the switch statement
	for {
		cmd := gl.GetInput()
		switch cmd[0] {
		case "quit":
			fmt.Println("Saving pokedex...")
			err = cfg.UpdateTeam(core.NewPokemon)
			if err != nil {
				fmt.Println("Error saving pokedex: ", err)
			}
			fmt.Println("Exiting P4T...")
			break repl
		case "catch":
			if cmd[1] == "-r" {
				poke, err := core.CatchRandom(cfg)
				if err != nil {
					fmt.Println(err)
					continue
				}
				s, e := gl.PokeToTypeColor(poke)
				text := gl.StringGradient(poke.Name, s, e)
				fmt.Println("CAUGHT: ", text)
				continue
			}
			poke, err := core.Catch(cfg, cmd[1])
			if err != nil {
				fmt.Printf("Error catching %s: %v\n", cmd[1], err)
				continue
			}
			s, e := gl.PokeToTypeColor(poke)
			text := gl.StringGradient(poke.Name, s, e)
			fmt.Println("CAUGHT: ", text)

		case "reset":
			fmt.Println("About to reset user team, are you sure? [y/n]")
			confirm := gl.GetInput()
			if confirm[0] != "y" && confirm[0] != "yes" {
				continue
			}
			fmt.Println("Okay - deleting users pokedex - goodluck on your new journey!")
			cfg.Reset(user[0])
			core.PokeDex = make(map[string]*gl.PokemonPlusCount)
			core.NewPokemon = make(map[string]int)
		case "pokedex":
			gl.DisplayPokedex(core.PokeDex)
		case "gui":
			fmt.Println("Starting p4r graphical user interface")
			if len(cmd) < 3 {
				fmt.Println("Too few arguments for gui, need a width and height")
				continue
			}
			w, err := strconv.Atoi(cmd[1])
			if err != nil {
				fmt.Println("Could not convert dimensions to int")
				continue
			}
			h, err := strconv.Atoi(cmd[2])
			if err != nil {
				fmt.Println("Could not convert dimensions to int")
				continue
			}
			ebiten.SetWindowSize(w, h)
			title := fmt.Sprintf("P4T - %s", user[0])
			ebiten.SetWindowTitle(title)
			TitleFont, err := gui.LoadFontFace("assets/fonts/PokeHollow.ttf", 72)
			if err != nil {
				fmt.Println("Could not load title font")
				continue
			}
			CommandFont, err := gui.LoadFontFace("assets/fonts/warownia-narrow.otf", 36)
			if err != nil {
				fmt.Println("Could not load command font")
				continue
			}
			commandRegistry := make(map[string]func(*gui.Game, ...string) error)
			commandRegistry["quit"] = gui.Quit
			commandRegistry["catch"] = gui.Catch

			g := &gui.Game{
				C:           core,
				Cfg:         &cfg,
				TitleFont:   TitleFont,
				CommandFont: CommandFont,
				User:        user[0],
				Commands:    commandRegistry,
				NextOpenSpot: gui.Vector{
					X: 0,
					Y: 100,
				},
			}
			err = ebiten.RunGame(g)
			if err != nil {
				fmt.Println("Error starting up gui: ", err)
			}
		default:
			fmt.Println("Error: Unrecognized Command")
		}
	}

}
