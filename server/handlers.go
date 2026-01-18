package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

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
	}

	var params Params
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Failed to decode json: %v\n", err)
		return
	}
	//This function is here to validate logins, so we need the auth package
}
