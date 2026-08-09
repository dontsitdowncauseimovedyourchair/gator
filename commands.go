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

func handlerReset(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("reset expects no arguments")
	}

	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("Users have been reset.")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("users expects no arguments")
	}

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		fmt.Printf("* %s", user.Name)
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf(" (current)")
		}
		fmt.Printf("\n")
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("agg expects no arguments")
	}

	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	fmt.Println(*feed)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}

	feedName := cmd.args[0]
	url := cmd.args[1]

	username := s.cfg.CurrentUserName

	user, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user %s does not exist", username)
		}
		return fmt.Errorf("%w", err)
	}

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("flop creating feed: %w", err)
	}
	fmt.Printf("Created feed %s at %s for user %s\n", feed.Name, feed.Url, username)
	fmt.Println(feed)
	return nil
}
func handlerFeeds(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("feeds expects no arguments")
	}

	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("flop fetching feeds: %w", err)
	}

	fmt.Println("Feeds:")
	for _, feed := range feeds {
		creator, err := s.db.GetFeedCreatorName(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("flop fetching feed creator name: %w", err)
		}
		fmt.Printf(" - \"%s\" created by %s from %s\n", feed.Name, creator, feed.Url)
	}
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
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", handlerAddFeed)
	commands.register("feeds", handlerFeeds)
	return commands
}
