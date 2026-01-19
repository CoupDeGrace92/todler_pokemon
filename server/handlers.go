package main

import (
	"coupdegrace92/pokemon_for_todlers/auth"
	"coupdegrace92/pokemon_for_todlers/server/database"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func (cfg *apiConfig) HandlerNewUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	var params parameters

	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding params: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Password stuff would be here - check if empty, hash it and check against db
	if params.Username == "" {
		log.Printf("Error: did not recieve a username\n")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := cfg.db.CreateUser(r.Context(), params.Username)
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

	//Here is a superficial solution that is not good
	if os.Getenv("PLATFORM") != "dev" {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Unauthorized user attempted to reset users")
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
