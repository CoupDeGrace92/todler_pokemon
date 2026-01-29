package main

import (
	"context"
	"coupdegrace92/pokemon_for_todlers/auth"
	"coupdegrace92/pokemon_for_todlers/server/database"
	"coupdegrace92/pokemon_for_todlers/shared"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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
	Xp      int `json:"base_experience"`
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

	fmt.Println("Checking Cache")
	exist := cfg.updateCache("pokeAPI", time.Minute*30)
	if exist {
		c := cfg.Caches["pokeAPI"]
		c.Mu.Lock()
		entry, ok := c.CacheItems[strId]
		c.Mu.Unlock()
		if ok {
			poke := entry.Val
			var p Poke
			err := json.Unmarshal(poke, &p)
			if err != nil {
				log.Println(err)
				return err
			}

			params := database.AddPokemonParams{
				ID:     int32(id),
				Name:   p.Name,
				Sprite: p.Sprites.Front,
				Type:   typesToString(p.Types),
				Url:    urlBase,
				BaseXp: int32(p.Xp),
			}
			fmt.Println("adding pokemon from cache")
			err = cfg.db.AddPokemon(context.Background(), params)
			if err != nil {
				log.Printf("Error adding pokemon %v to db: %v\n", params.Name, err)
				return err
			}
			return nil
		}
	}

	fmt.Println("getting pokemon from pokeapi")
	resp, err := http.Get(urlBase)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		err = fmt.Errorf("NOT FOUND")
		fmt.Println(err)
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

	byteSlice, err := json.Marshal(p)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Println("Adding pokemon to cache")
	cfg.Caches["pokeAPI"].Add(strId, byteSlice)

	params := database.AddPokemonParams{
		ID:     int32(id),
		Name:   p.Name,
		Sprite: p.Sprites.Front,
		Type:   typesToString(p.Types),
		Url:    urlBase,
		BaseXp: int32(p.Xp),
	}
	fmt.Println("Adding pokemon to db")
	err = cfg.db.AddPokemon(context.Background(), params)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (cfg *apiConfig) HandlerPopulatePokeDB(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Attempting to populate pokemon db")
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
	fmt.Println(count)

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
		fmt.Println(i)
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

func (cfg *apiConfig) updateCache(name string, interval time.Duration) bool {
	_, ok := cfg.Caches[name]
	if !ok {
		cache := shared.NewCache(interval)
		cfg.Caches[name] = cache
		return false
	}
	return true
}

func (cfg *apiConfig) HandlerAddToTeam(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Error getting bearer token: ", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.Secret)
	if err != nil {
		log.Printf("Error validating token: %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	username, err := cfg.db.GetUserFromID(r.Context(), userID)
	if err != nil {
		log.Println("Error grabbing username from ID: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	poke := make(map[string]int)
	err = json.NewDecoder(r.Body).Decode(&poke)
	if err != nil {

	}
	for poke, count := range poke {
		poke := strings.ToLower(poke)
		dbcount, err := cfg.db.ValidatePokemon(r.Context(), poke)
		if err != nil {
			log.Printf("Error finding %s in database: %v\n", poke, err)
			continue
		} else if dbcount == 0 {
			log.Printf("%s is not a pokemon in the database\n", poke)
			continue
		}

		params := database.AddPokemonToTeamParams{
			UserName: username,
			Poke:     poke,
			Count:    int32(count),
		}

		err = cfg.db.AddPokemonToTeam(r.Context(), params)
		if err != nil {
			log.Printf("Error adding %s to %s's team: %v \n", params.Poke, params.UserName, err)
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Team added to database"))
}

func (cfg *apiConfig) HandlerGetTeam(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Println("Error getting bearer jwt token: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.Secret)
	if err != nil {
		log.Println("Error validating jwt token: ", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := cfg.db.GetUserFromID(r.Context(), userID)
	if err != nil {
		log.Println("Error getting user from db: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	team, err := cfg.db.GetTeam(r.Context(), user)
	if err != nil {
		log.Println("Error getting team from db: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type P struct {
		Id     int32  `json:"id"`
		Name   string `json:"name"`
		Sprite string `json:"sprite_url"`
		Types  string `json:"types"`
		Url    string `json:"url"`
	}

	type Team struct {
		Pokemon   []P       `json:"pokemon"`
		FetchedAt time.Time `json:"fetched_at"`
	}

	var pokemon []P
	fetch := time.Now()
	outTeam := Team{
		Pokemon:   pokemon,
		FetchedAt: fetch,
	}
	for _, p := range team {
		pObject, err := cfg.db.GetPokeByName(r.Context(), p.Poke)
		if err != nil {
			log.Printf("Error getting pokemon entry for %v: %v\n", p.Poke, err)
			continue
		}

		cleanedPoke := P{
			Id:     pObject.ID,
			Name:   pObject.Name,
			Sprite: pObject.Sprite,
			Types:  pObject.Type,
			Url:    pObject.Url,
		}
		outTeam.Pokemon = append(outTeam.Pokemon, cleanedPoke)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(outTeam)
	if err != nil {
		log.Println("Error encoding team into a json: ", err)
	}
}

func (cfg *apiConfig) HandlerGetPokemon(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Println("Error getting token from request: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = auth.ValidateJWT(token, cfg.Secret)
	if err != nil {
		log.Println("Unauthourized request: ", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	cTypeHeader := r.Header["Content-Type"]
	if len(cTypeHeader) == 0 {
		log.Println("No Content-Type header")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please include Content-Type header specifying text/plain"))
		return
	}
	parts := strings.Fields(cTypeHeader[0])
	cType := strings.TrimSuffix(parts[0], ";")
	if cType != "text/plain" {
		log.Println("Content-Type not text/plain")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please send only Content-Type: text/plain to this endpoint"))
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Can not read request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error: Can not read request"))
		return
	}

	pokemon := string(bodyBytes)
	pokeID, err := strconv.Atoi(pokemon)
	if err != nil {
		p, err := cfg.db.GetPokeByName(r.Context(), pokemon)
		if err != nil {
			log.Println("Error getting pokemon from db")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Error retrieving %s from the database", pokemon)))
			return
		}

		type resp struct {
			Id     int32  `json:"id"`
			Name   string `json:"name"`
			Sprite string `json:"sprite_url"`
			Types  string `json:"types"`
			Url    string `json:"pokeapi_url"`
			BaseXp int32  `json:"base_xp"`
		}

		finalResp := resp{
			Id:     p.ID,
			Name:   p.Name,
			Sprite: p.Sprite,
			Types:  p.Type,
			Url:    p.Url,
			BaseXp: p.BaseXp,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(finalResp)
		if err != nil {
			log.Println("Error encoding pokeresponse struct")
		}
		return
	}
	p, err := cfg.db.GetPokeByID(r.Context(), int32(pokeID))
	if err != nil {
		log.Println("Error getting pokemon from db")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Error retrieving %s from the database", pokemon)))
		return
	}

	type resp struct {
		Id     int32  `json:"id"`
		Name   string `json:"name"`
		Sprite string `json:"sprite_url"`
		Types  string `json:"types"`
		Url    string `json:"pokeapi_url"`
		BaseXp int32  `json:"base_xp"`
	}

	finalResp := resp{
		Id:     p.ID,
		Name:   p.Name,
		Sprite: p.Sprite,
		Types:  p.Type,
		Url:    p.Url,
		BaseXp: p.BaseXp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(finalResp)
	if err != nil {
		log.Println("Error encoding pokeresponse struct")
	}
}

func (cfg *apiConfig) HandlerResetTeams(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("PLATFORM") != "dev" {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("Unauthorized user tried to reset teams db")
		return
	}

	err := cfg.db.ResetTeams(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error resetting teams db: ", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pokemon teams db reset"))
}

func (cfg *apiConfig) HandlerResetUserTeams(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("Unauthourized attempt to delete a users pokemon")
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.Secret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("Unauthourized attempt to delete a users pokemon")
		return
	}

	username, err := cfg.db.GetUserFromID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error retrieving username from db for userid: ", userID)
		return
	}

	user := r.PathValue("user")
	if username != user {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = cfg.db.ResetUserTeam(r.Context(), user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error resetting %s's team\n", user)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Succesfully reset users caught pokemon"))
}

func (cfg *apiConfig) HandlerGetWeightedIds(w http.ResponseWriter, r *http.Request) {
	exists := cfg.updateCache("index", time.Hour)
	cfg.Caches["index"].Mu.Lock()
	defer cfg.Caches["index"].Mu.Unlock()
	var index []int
	if exists {
		c := cfg.Caches["index"]
		if wIndex, ok := c.CacheItems["weightedIndex"]; ok {
			rawCSVStr := string(wIndex.Val)
			strSlice := strings.Fields(rawCSVStr)
			var rawInts []int
			for _, i := range strSlice {
				x, err := strconv.Atoi(strings.TrimSuffix(i, ","))
				if err != nil {
					log.Println("Error converting string to int: ", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				rawInts = append(rawInts, x)
			}
			index = rawInts
		}

	}
	//We need to check the existance of "weightedIndex" subCaches
	c := cfg.Caches["index"]
	_, ok := c.CacheItems["weightedIndex"]
	if !ok {
		c.CacheItems["weightedIndex"] = &shared.CacheEntry{}
	}

	if index == nil {
		last := 0
		count, err := cfg.db.NumPokemon(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Error getting pokemon count")
			return
		}
		for i := 1; i <= int(count); i++ {
			poke, err := cfg.db.GetPokeByID(r.Context(), int32(i))
			if err != nil {
				log.Printf("Error getting pokemon with id %v from the db\n", i)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			xp := poke.BaseXp
			if xp > 400 {
				xp = 400
			}
			normalized := int((410 - xp) / 4)
			inclusiveEnd := last + normalized
			last = inclusiveEnd
			index = append(index, inclusiveEnd)
		}
		//Now we want to add those values to the cache
		//The weighted index needs to be converted into a comma delimated format with spaces
		cacheIndexStr := ""
		for _, x := range index {
			strx := strconv.Itoa(x)
			modstr := strx + ", "
			cacheIndexStr += modstr
		}

		wIndexCache := c.CacheItems["weightedIndex"]
		wIndexCache.Val = []byte(cacheIndexStr)
		wIndexCache.CreatedAt = time.Now()
	}

	type Resp struct {
		WeightedEnds []int
	}
	resp := Resp{
		WeightedEnds: index,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
