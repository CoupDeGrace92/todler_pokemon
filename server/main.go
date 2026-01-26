package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"coupdegrace92/pokemon_for_todlers/server/database"
	"coupdegrace92/pokemon_for_todlers/shared"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	db     *database.Queries
	Secret string
	Caches map[string]*shared.Cache
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
	Refresh   string    `json:"refresh_token"`
	Token     string    `json:"token"`
	Password  string    `json:"password"`
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
	caches := make(map[string]*shared.Cache)
	apiCfg := apiConfig{
		db:     dbQueries,
		Secret: os.Getenv("JWT_SECRET"),
		Caches: caches,
	}

	ServerMux := http.NewServeMux()

	ServerMux.Handle("POST /api/register", http.HandlerFunc(apiCfg.HandlerNewUser))
	ServerMux.Handle("DELETE /admin/reset", http.HandlerFunc(apiCfg.HandlerReset))
	ServerMux.Handle("POST /api/login", http.HandlerFunc(apiCfg.HandlerLogin))
	ServerMux.Handle("POST /admin/poke/populate", http.HandlerFunc(apiCfg.HandlerPopulatePokeDB))
	ServerMux.Handle("DELETE /admin/poke/reset", http.HandlerFunc(apiCfg.HandlerResetPokemon))

	server := &http.Server{
		Handler: ServerMux,
		Addr:    os.Getenv("PORT"),
	}

	fmt.Println("Serving on port ", server.Addr)

	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}

}
