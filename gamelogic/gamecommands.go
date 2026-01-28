package gamelogic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
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
	fmt.Println("reset              |None                    |Resets caught pokemon to none     ")

	fmt.Println()
	fmt.Println()
}

func CatchRandom() (id int) {
	//Generate weighted indicies
	//Get last num, slice of first indicies
	// rand.Intn(last num)
	defer fmt.Println("> ")
	rawurl := os.Getenv("SERVER_URL")
	fmt.Println(rawurl)
	url := rawurl + "/api/pokemon/weights"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error getting weights table")
		return
	}

	req.Header.Add("User-Agent", "Go-http-client")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error with http request ", err)
		return
	}
	defer resp.Body.Close()
	if resp.Status != "200 OK" {
		log.Println("Server status: ", resp.Status)
		return
	}

	type respJson struct {
		Weights []int `json:"WeightedEnds"`
	}
	var j respJson
	err = json.NewDecoder(resp.Body).Decode(&j)
	if err != nil {
		log.Printf("Error decoding weights json into local struct: %v\n", err)
		return
	}
	weightedIndex := j.Weights
	if len(weightedIndex) == 0 {
		fmt.Println("Error: Weights list is nil")
		return 0
	}
	r := rand.Intn(weightedIndex[len(weightedIndex)-1])
	ID := sort.Search(len(weightedIndex), func(i int) bool {
		return weightedIndex[i] >= r
	}) + 1
	//Then search the int slice for the subset in which the rand resides
	//first, we have to update the db to include base xp values, update queries to return those as well, and update handlers
	return ID
}

func UpdateEnvFile(path, key, value string) error {
	envMap, err := godotenv.Read(path)
	if err != nil {
		if os.IsNotExist(err) {
			envMap = map[string]string{}
		} else {
			return err
		}
	}

	envMap[key] = value
	return godotenv.Write(envMap, path)
}

func Register(user, pass string) error {
	rawurl := os.Getenv("SERVER_URL")
	url := rawurl + "/api/register"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	//WE HAVE TO BUILD THE BODY HERE
	type Request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	r := Request{
		Username: user,
		Password: pass,
	}
	rBytes, err := json.Marshal(r)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(rBytes))
	if err != nil {
		log.Printf("Error creating request")
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error getting response from client for registration req: ", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Returned status not okay: ", resp.Status)
		return nil
	}
	return nil
}

func Login(user, pass string) error {
	rawurl := os.Getenv("SERVER_URL")
	url := rawurl + "/api/login"
	client := &http.Client{
		Timeout: time.Second * 20,
	}
	type Request struct {
		Username string `json:"username"`
		Pass     string `json:"password"`
	}
	r := Request{
		Username: user,
		Pass:     pass,
	}

	rBytes, err := json.Marshal(r)
	if err != nil {
		fmt.Println("Error marshalling json: ", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(rBytes))
	if err != nil {
		fmt.Println("Error creating request: ", err)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error getting response from server: ", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Response status not okay: ", resp.Status)
		return nil
	}

	type Tokens struct {
		JWT     string `json:"token"`
		Refresh string `json:"refresh_token"`
	}
	var t Tokens
	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		fmt.Println("Error decoding token jsons: ", err)
		return err
	}
	err = UpdateEnvFile("./client/.env", "JWT", t.JWT)
	if err != nil {
		fmt.Println("Error moving JWT to env file: ", err)
	}
	err = UpdateEnvFile("./client/.env", "REFRESH_TOKEN", t.Refresh)
	if err != nil {
		fmt.Println("Error moving refresh to env file: ", err)
		return err
	}

	return nil
}

func GetPokemon(id int) error {
	return nil
}
