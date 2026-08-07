package main

import (
	"log"
	"os"

	"github.com/dontsitdowncauseimovedyourchair/gator/internal/config"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatalf("Not enough arguments provided")
	}

	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("couldn't read config: %v", err)
	}

	globalState := state{
		cfg: &cfg,
	}

	commands := commands{handlers: make(map[string]func(*state, command) error)}
	commands.register("login", handlerLogin)

	userCommand := os.Args[1]
	userArgs := os.Args[2:]

	cmd := command{
		name: userCommand,
		args: userArgs,
	}

	if err := commands.run(&globalState, cmd); err != nil {
		log.Fatalf("%s flop: %v", cmd.name, err)
	}

}
