package main

import (
	pApi "coupdegrace92/pokemon_for_todlers/pokeapi"
	"fmt"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Problem loading .env: ", err)
		return
	}
	fmt.Println("Welcome to Pokemon for Todlers (and dad)!")
	poke, err := pApi.GetPokemon("pikachu")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%#v\n", poke)

	err = pApi.SpriteToFile("pikachu")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Created png for pikachu")
	}
}
