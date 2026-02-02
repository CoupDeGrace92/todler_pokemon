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
	"strconv"
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
	fmt.Println("__________________________________________________________________________________")
	fmt.Println("catch              |-r:  random mode        |Will catch a pokemon, a random one")
	fmt.Println("                   |<pokemon>:              |if -r is specified, else <pokemon>")
	fmt.Println("pokedex            |None                    |Will display pokemon caught and count")
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

func Register(user, pass string) (outerror error, success bool) {
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
		return err, false
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error getting response from client for registration req: ", err)
		return err, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		log.Println("Returned status not okay: ", resp.Status)
		return nil, false
	}
	return nil, true
}

func Login(user, pass string) (outerror error, success bool) {
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
		return err, false
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(rBytes))
	if err != nil {
		fmt.Println("Error creating request: ", err)
		return err, false
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error getting response from server: ", err)
		return err, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Response status not okay: ", resp.Status)
		return nil, false
	}

	type Tokens struct {
		JWT     string `json:"token"`
		Refresh string `json:"refresh_token"`
	}
	var t Tokens
	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		fmt.Println("Error decoding token jsons: ", err)
		return err, false
	}
	err = UpdateEnvFile("./client/.env", "JWT", t.JWT)
	if err != nil {
		fmt.Println("Error moving JWT to env file: ", err)
		return err, false
	}
	err = UpdateEnvFile("./client/.env", "REFRESH_TOKEN", t.Refresh)
	if err != nil {
		fmt.Println("Error moving refresh to env file: ", err)
		return err, false
	}

	return nil, true
}

func GetPokemon(id string) (Pokemon, error) {
	rawurl := os.Getenv("SERVER_URL")
	url := rawurl + "/api/pokemon"
	client := &http.Client{
		Timeout: time.Second * 20,
	}

	body := []byte(id)

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))
	if err != nil {
		x := fmt.Sprintf("Error generating request: %s\n", err)
		fmt.Print(x)
		err = fmt.Errorf(x)
		return Pokemon{}, err
	}
	token := os.Getenv("JWT")
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error getting server response")
		return Pokemon{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Error: Status code not ok: %v\n", resp.Status)
		return Pokemon{}, err
	}

	var poke RecievedPokemon
	if err = json.NewDecoder(resp.Body).Decode(&poke); err != nil {
		fmt.Println("Error decoding json: ", err)
		return Pokemon{}, err
	}
	//TYPE LIST HERE - csv delimated, use fields strip trailing comma
	var typeList []string
	uncleanedList := strings.Fields(poke.Types)
	for _, j := range uncleanedList {
		cleaned := strings.TrimLeft(j, ",")
		typeList = append(typeList, cleaned)
	}
	//SPRITE LIST HERE
	spriteList := make(map[string]string)
	//currently there is only a single string that goes in here - we will modify when we include back sprite
	spriteList["front"] = poke.Sprites

	cleanedPoke := Pokemon{
		Id:      int(poke.Id),
		Name:    poke.Name,
		Types:   typeList,
		Sprites: spriteList,
		Url:     poke.Url,
		Xp:      int(poke.Xp),
	}

	return cleanedPoke, nil
}

func AddToMap(name string, pokeMap map[string]int) {
	count, ok := pokeMap[name]
	if !ok {
		pokeMap[name] = 1
		return
	}
	pokeMap[name] = count + 1
}

func Reset(user string) {
	rawurl := os.Getenv("SERVER_URL")
	url := rawurl + "/api/teams/" + user
	client := http.Client{
		Timeout: time.Second * 20,
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Println("Error resetting users teams: ", err)
		return
	}

	token := os.Getenv("JWT")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request: ", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("Error: status was not OK: ", resp.Status)
		return
	}
}

func DisplayPokedex(pokedex map[string]int) {
	//Thinking this through - I want to create a list of all pokemon in the pokedex, then sort it alphabetically
	var pokeList []string
	for i := range pokedex {
		pokeList = append(pokeList, i)
	}
	sort.Strings(pokeList)
	fmt.Println("\033[31m___________________________________________________________\033[0m")
	fmt.Println("Pokemon\033[31m                       |\033[0mcount\033[31m                       \033[0m")
	fmt.Println("\033[31m___________________________________________________________\033[0m")
	//30 characters left of |
	for _, poke := range pokeList {
		spaces := ""
		num := len(poke)
		for i := 0; i < 30-num; i++ {
			spaces += " "
		}
		fmt.Printf("%s%s\033[31m|\033[0m%v\n", poke, spaces, pokedex[poke])
	}
}

func GetTeam() ([]RecievedPokemon, error) {
	rawurl := os.Getenv("SERVER_URL")
	url := rawurl + "/api/teams"
	client := http.Client{
		Timeout: time.Second * 20,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating request ", err)
		return nil, err
	}
	token := os.Getenv("JWT")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request ", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("Response status not OK: ", resp.Status)
		return nil, err
	}

	type Team struct {
		Pokemon   []RecievedPokemon `json:"pokemon"`
		FetchedAt time.Time         `json:"fetched_at"`
	}

	var rawOut Team
	if err = json.NewDecoder(resp.Body).Decode(&rawOut); err != nil {
		fmt.Println("Error decoding JSON: ", err)
		return nil, err
	}
	out := rawOut.Pokemon
	return out, nil
}

func UpdateTeam(pokedex map[string]int) error {
	rawurl := os.Getenv("SERVER_URL")
	url := rawurl + "/api/teams"
	client := http.Client{
		Timeout: time.Second * 20,
	}
	rawBytes, err := json.Marshal(pokedex)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(rawBytes))
	if err != nil {
		log.Println("Error creating request ", err)
		return err
	}
	token := os.Getenv("JWT")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error processing request: ", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Response status not OK: ", resp.Status)
	}
	return nil
}

func TripToInts(s string) ([]int, error) {
	s = strings.TrimPrefix(s, "(")
	s = strings.TrimSuffix(s, ")")
	strSlice := strings.Split(s, ",")
	var outSlice []int
	for _, i := range strSlice {
		num, err := strconv.Atoi(i)
		if err != nil {
			return nil, err
		}
		outSlice = append(outSlice, num)
	}
	return outSlice, nil
}

func AddZeros(i int) string {
	s := strconv.Itoa(i)
	if len(s) > 3 {
		return ""
	}
	if len(s) < 3 {
		for i := 0; i < 3-len(s); i++ {
			s = "0" + s
		}
	}
	return s
}

func ColorGradient(startColors []int, endColors []int, interval int) ([]int, []int, []int) {
	var reds []int
	var greens []int
	var blues []int
	if interval <= 0 {
		return reds, greens, blues
	}
	if min(startColors[0], startColors[1], startColors[2], endColors[0], endColors[1], endColors[2]) < 0 {
		return reds, greens, blues
	} else if max(startColors[0], startColors[1], startColors[2], endColors[0], endColors[1], endColors[2]) > 255 {
		return reds, greens, blues
	}
	redDif := (endColors[0] - startColors[0]) / (interval - 1)
	greenDif := (endColors[1] - startColors[1]) / (interval - 1)
	blueDif := (endColors[2] - startColors[2]) / (interval - 1)
	for i := 0; i < interval; i++ {
		r := startColors[0] + i*redDif
		b := startColors[2] + i*blueDif
		g := startColors[1] + i*greenDif
		reds = append(reds, r)
		blues = append(blues, b)
		greens = append(greens, g)
	}
	return reds, greens, blues
}

func StringGradient(s string, rgbStart, rgbEnd []int) string {
	interval := len(s)
	reds, greens, blues := ColorGradient(rgbStart, rgbEnd, interval)
	var outString strings.Builder
	for i, char := range s {
		red := AddZeros(reds[i])
		green := AddZeros(greens[i])
		blue := AddZeros(blues[i])
		c := fmt.Sprintf("\033[38;2;%s;%s;%sm%c\033[0m", red, green, blue, char)
		outString.WriteString(c)
	}
	return outString.String()
}
