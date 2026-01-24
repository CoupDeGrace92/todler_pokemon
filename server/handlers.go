package main

import (
	"context"
	"coupdegrace92/pokemon_for_todlers/auth"
	"coupdegrace92/pokemon_for_todlers/server/database"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

func (cfg *apiConfig) HandlerNewUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	var rawParams parameters

	err := decoder.Decode(&rawParams)
	if err != nil {
		log.Printf("Error decoding params: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hash, err := auth.HashPassword(rawParams.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error getting password hash: %v", err)
		return
	}

	if rawParams.Username == "" {
		log.Printf("Error: did not recieve a username\n")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userParams := database.CreateUserParams{
		UserName: rawParams.Username,
		PassHash: hash,
	}
	user, err := cfg.db.CreateUser(r.Context(), userParams)
	if err != nil {
		log.Printf("Error creating user: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Username:  user.UserName,
	}
	dat, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling json: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	w.Write(dat)
}

func (cfg *apiConfig) HandlerReset(w http.ResponseWriter, r *http.Request) {
	//We need to add admin access only here - verify user

	//Here is a superficial solution that is not good - used for scaffolding, remove for prod
	if os.Getenv("PLATFORM") != "dev" {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("Unauthorized user attempted to reset users")
		return
	}

	err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error resetting db: %v\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Database reset"))
}

func (cfg *apiConfig) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	type Params struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Expires  int    `json:"expires_in"`
	}

	var params Params
	params.Expires = 3600

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Failed to decode json: %v\n", err)
		return
	}
	if params.Expires > 3600 {
		params.Expires = 3600
	}

	user, err := cfg.db.GetUser(r.Context(), params.Username)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Failed to decode json: %v\n", err)
		return
	}

	authenticated, err := auth.CheckPasswordHash(params.Password, user.PassHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error checking pass hash: %v\n", err)
		return
	}
	if !authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("Password did not match for login attempt on ", user.UserName)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.Secret, time.Duration(params.Expires)*time.Second)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error making JWT for user: %v\n", err)
		return
	}

	refresh, err := auth.MakeRefreshToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error making refresh for user: %v\n", err)
		return
	}

	outUser := User{
		ID:       user.ID,
		Username: user.UserName,
		Token:    token,
		Refresh:  refresh,
	}

	args := database.InsertRefreshParams{
		Token:     refresh,
		UserName:  user.UserName,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	}
	err = cfg.db.InsertRefresh(r.Context(), args)
	if err != nil {
		log.Printf("Error adding refresh token to db: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(outUser)
	if err != nil {
		log.Printf("Error encoding user struct back to user: %s", err)
	}
}

// The following structs are for filling our local db from pokeAPI
type Poke struct {
	Id      int         `json:"id"`
	Name    string      `json:"name"`
	Types   []PokeType  `json:"types"`
	Sprites PokeSprites `json:"sprites"`
	Url     string
}

type PokeType struct {
	Type map[string]interface{} `json:"type"`
}

type PokeSprites struct {
	Front string `json:"front_default"`
	Back  string `json:"back_default"`
	//We can get older sprites in the Poke struct
	//The JSON classification is generations and not sprites
}

func getPokeAPICount() (int, error) {
	countUrl := "https://pokeapi.co/api/v2/pokemon/"

	resp, err := http.Get(countUrl)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Response status not OK: ", resp.StatusCode)
		return 0, err
	}

	countOut := make(map[string]interface{})
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&countOut); err != nil {
		log.Println(err)
		return 0, err
	}

	count, ok := countOut["count"].(float64)
	if !ok {
		log.Printf("Count does not have the type expected, expected int, got: %T\n", countOut["count"])
		return 0, err
	}
	out := int(count)
	return out, nil
}

func typesToString(t []PokeType) string {
	out := ""
	for i, x := range t {
		t, ok := x.Type["name"].(string)
		if !ok {
			log.Println("Issue converting pokemon type")
			continue
		}
		if i == 0 {
			out = t
		} else {
			out = out + ", " + t
		}
	}
	return out
}

func (cfg *apiConfig) addPokeToDB(id int) error {
	strId := strconv.Itoa(id)
	urlBase := "http://pokeapi.co/api/v2/pokemon/" + strId
	resp, err := http.Get(urlBase)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		err = fmt.Errorf("NOT FOUND")
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Println("Response status not ok: ", resp.StatusCode)
		err = fmt.Errorf("Response status not ok: %v", resp.StatusCode)
		return err
	}

	var p Poke
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&p); err != nil {
		log.Printf("Error decoding response: %v\n", err)
		return err
	}

	params := database.AddPokemonParams{
		ID:     int32(id),
		Name:   p.Name,
		Sprite: p.Sprites.Front,
		Type:   typesToString(p.Types),
		Url:    urlBase,
	}

	err = cfg.db.AddPokemon(context.Background(), params)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (cfg *apiConfig) HandlerPopulatePokeDB(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("PLATFORM") != "dev" {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("Unauthorized user tried to populate pokemon db")
		return
	}

	count, err := getPokeAPICount()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error getting the count of pokemon: ", err)
		return
	}

	for i := 1; i <= count; i++ {
		err = cfg.addPokeToDB(i)
		if err != nil && err != fmt.Errorf("NOT FOUND") {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		} else if err == fmt.Errorf("NOT FOUND") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Pokemon db populated, max pokemon: %v", i-1)))
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pokemon DB populated"))
}

func (cfg *apiConfig) HandlerResetPokemon(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("PLATFORM") != "dev" {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("Unauthorized user tried to reset pokemon db")
		return
	}

	err := cfg.db.ResetPokemon(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error resetting pokemon db: ", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pokemon DB reset"))
}
