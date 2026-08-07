package main

import (
	"fmt"
	"log"

	"github.com/dontsitdowncauseimovedyourchair/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("couldn't read config: %v", err)
	}

	err = cfg.SetUser("Alex")
	if err != nil {
		log.Fatalf("flopped reading config: %v", err)
	}

	cfg, err = config.Read()
	if err != nil {
		log.Fatalf("couldn't read config: %v", err)
	}
	fmt.Println(cfg)
}