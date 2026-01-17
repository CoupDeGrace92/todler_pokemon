package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"coupdegrace92/pokemon_for_todlers/server/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	db *database.Queries
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Problem loading .env: ", err)
		return
	}
	fmt.Println("Spinning up server")

	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Unexpected error opening database: %v\n", err)
		return
	}

	dbQueries := database.New(db)

	//OUR HANDLERS ARE EMPTY - THIS CONFIG WILL BE IMPORTANT FOR DB INTERACTION
	apiCfg := apiConfig{
		db: dbQueries,
	}

	ServerMux := http.NewServeMux()

	ServerMux.Handle("/api/register", http.HandlerFunc(apiCfg.HandlerNewUser))

	server := &http.Server{
		Handler: ServerMux,
		Addr:    os.Getenv("PORT"),
	}

	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}

}
