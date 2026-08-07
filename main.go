package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/dontsitdowncauseimovedyourchair/gator/internal/config"
	"github.com/dontsitdowncauseimovedyourchair/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatalf("Not enough arguments provided")
	}

	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("couldn't read config: %v", err)
	}

	dbURL := "postgres://yucateco:@localhost:5432/gator"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("unable to open database at %s: %v", dbURL, err)
	}
	dbQueries := database.New(db)

	globalState := state{
		db:  dbQueries,
		cfg: &cfg,
	}

	userCommand := os.Args[1]
	userArgs := os.Args[2:]

	cmd := command{
		name: userCommand,
		args: userArgs,
	}

	if err := commands.run(&globalState, cmd); err != nil {
		log.Fatalf("%v", err)
	}

}
