package main

import (
	"encoding/json"
	"log"
	"net/http"

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
