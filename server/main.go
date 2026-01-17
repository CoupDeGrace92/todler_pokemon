package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"coupdegrace92/pokemon_for_todlers/server/database"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	db *database.Queries
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
	apiCfg = apiConfig{
		db: dbQueries,
	}

	ServerMux := http.NewServeMux()

	/* CURRENTLY WE DON'T HAVE ANYTHING TO SERVE WITH THE FILE SERVER
	fileserverPath := http.Dir("./assets/")
	fileserver := http.FileServer(fileserverPath)
	*/

	//Register handles to the server mux here

	server := &http.Server{
		Handler: ServerMux,
		Addr:    os.Getenv("PORT"),
	}

	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}

}
