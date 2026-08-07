package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dontsitdowncauseimovedyourchair/gator/internal/database"
	"github.com/google/uuid"
)

type command struct {
	name string
	args []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: login <username>")
	}
	username := cmd.args[0]

	user, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user %s does not exist", username)
		}
		return fmt.Errorf("%w", err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("Logged in as %s\n", username)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: register <username>")
	}

	username := cmd.args[0]
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})
	if err != nil {
		return fmt.Errorf("flop registering user: %w", err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("user with name %s was successfully created and logged in!\n", user.Name)
	fmt.Printf("user struct: %v\n", user)
	return nil
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	if s == nil {
		return fmt.Errorf("%s flop: no state provided", cmd.name)
	}
	if err := c.handlers[cmd.name](s, cmd); err != nil {
		return fmt.Errorf("%s flop: %w", cmd.name, err)
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func getCommands() commands {
	commands := commands{handlers: make(map[string]func(*state, command) error)}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	return commands
}
