package pApi

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"strings"
	"time"
)

type Pokemon struct {
	Name          string                 `json:"name"`
	ID            int                    `json:"id"`
	Sprites       Sprites                `json:"sprites"`
	BaseXp        int                    `json:"base_experience"`
	OlderVersions map[string]interface{} `json:"versions"`
}

type Sprites struct {
	Back  string `json:"back_default"`
	Front string `json:"front_default"`
}

func GetPokemon(name string) (*Pokemon, error) {
	//Name can be either its given name or an id
	PokemonEnd := os.Getenv("POKEMON_ENDPOINT")
	fmt.Print(PokemonEnd)
	client := http.Client{
		Timeout: time.Second * 10,
	}
	fmt.Println(PokemonEnd + name + "/")
	req, err := http.NewRequest("GET", PokemonEnd+name+"/", nil)
	if err != nil {
		fmt.Printf("Error creating get request: %v\n", err)
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error getting response from client: %v\n", err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Error: status code not 200: %v", resp.StatusCode)
		err = errors.New(msg)
		fmt.Println(err)
		return nil, err
	}

	var pokemon Pokemon
	err = json.NewDecoder(resp.Body).Decode(&pokemon)
	if err != nil {
		fmt.Printf("Error decoding json into pokemon struct: %s", err)
		return nil, err
	}

	return &pokemon, nil
}

func GetSprite(poke Pokemon, opts ...string) (img image.Image, format string, err error) {
	spriteFront := true
	for _, opt := range opts {
		if strings.ToLower(opt) == "back" {
			spriteFront = false
		}
	}
	var url string
	if spriteFront {
		url = poke.Sprites.Front
	} else {
		url = poke.Sprites.Back
	}
	client := http.Client{
		Timeout: time.Second * 30,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating http request: %v\n", err)
		return nil, "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error getting response from http request: %v\n", err)
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Response status not ok: %v\n", resp.StatusCode)
		return nil, "", err
	}

	outImg, format, err := image.Decode(resp.Body)
	if err != nil {
		fmt.Printf("Error decoding image from response: %v", err)
		return nil, "", err
	}
	fmt.Println("Image decoded with format: ", format)
	return outImg, format, nil
}

func PngImageToFile(img image.Image, format, filepath string) error {
	if format != "png" {
		fmt.Printf("Error: trying to write non-png to png")
		err := errors.New("Trying to write non-pgn to pgn")
		return err
	}
	f, err := os.Create(filepath)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return err
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		fmt.Printf("Error encoding png: %v", err)
		return err
	}
	return nil
}

func SpriteToFile(name string, opts ...string) error {
	poke, err := GetPokemon(name)
	if err != nil {
		return err
	}
	var keyOpt string
	for _, opt := range opts {
		if strings.ToLower(opt) == "back" {
			keyOpt = "back"
			break
		}
	}
	img, format, err := GetSprite(*poke, keyOpt)
	if err != nil {
		return err
	}
	if keyOpt == "" {
		keyOpt = "front"
	}
	filePath := fmt.Sprintf("assets/pokemon/images/%s_%s.png", poke.Name, keyOpt)
	err = PngImageToFile(img, format, filePath)
	if err != nil {
		return err
	}
	return nil
}

