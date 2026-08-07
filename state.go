package main

import (
	"github.com/dontsitdowncauseimovedyourchair/gator/internal/config"
	"github.com/dontsitdowncauseimovedyourchair/gator/internal/database"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}
