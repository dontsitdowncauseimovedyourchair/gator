package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
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
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage agg <interval>, interval examples: 1h | 5m | 10s | 500ms | 1h10m5s")
	}

	duration := cmd.args[0]
	interval, err := time.ParseDuration(duration)
	if err != nil {
		return fmt.Errorf("interval flop: %w", err)
	}
	if interval < time.Duration(time.Millisecond*100) {
		return fmt.Errorf("minimum aggregation interval is 100ms, avoid DDOS-ing pls\n")
	}

	fmt.Printf("Collecting feeds every %s\n", interval.String())

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		err = scrapeFeeds(context.Background(), s)
		if err != nil {
			fmt.Printf("feed flop: %s\n", err.Error())
		}
	}

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}

	feedName := cmd.args[0]
	url := cmd.args[1]

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

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to create feed follow: %w", err)
	}

	fmt.Printf("Created feed %s at %s for user %s\n", feed.Name, feed.Url, user.Name)
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

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: follow <url>")
	}

	url := cmd.args[0]
	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %w", err)
	}

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to create feed follow: %w", err)
	}

	fmt.Printf("Subscribed to feed \"%s\" as user %s\n", feed.Name, user.Name)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("following expects no arguments")
	}

	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("flop fetching feeds followed: %w", err)
	}

	if len(feeds) > 0 {
		fmt.Printf("Feeds %s is following:\n", user.Name)
		for i := range feeds {
			fmt.Printf(" - %s\n", feeds[i].FeedName)
		}
	} else {
		fmt.Printf("User %s is currently following no feeds.\n", user.Name)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: unfollow <url>")
	}

	url := cmd.args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %w", err)
	}

	err = s.db.DeleteFeedFollowRow(context.Background(), database.DeleteFeedFollowRowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("flopped unfollowing: %w", err)
	}

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2 //Default
	var err error

	if len(cmd.args) > 1 {
		return fmt.Errorf("usage: browse [limit]")
	} else if len(cmd.args) == 1 {
		limit, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("%s is not a number for posts limit", cmd.args[0])
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("flop fetching posts: %w", err)
	}

	for i := range posts {
		fmt.Printf(" - %s - %s\n\t%s\n\tLink: %s\n", posts[i].Title, posts[i].PublishedAt.Time, posts[i].Description.String, posts[i].Url)
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
	commandable, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command %s", cmd.name)
	}
	if err := commandable(s, cmd); err != nil {
		return fmt.Errorf("%s flop: %w", cmd.name, err)
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

// handle commands that require me to log in better-ly
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		username := s.cfg.CurrentUserName
		user, err := s.db.GetUser(context.Background(), username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("user %s does not exist", username)
			}
			return fmt.Errorf("flop fetching user: %w", err)
		}
		err = handler(s, cmd, user)
		if err != nil {
			return err
		}

		return nil
	}
}

func getCommands() commands {
	commands := commands{handlers: make(map[string]func(*state, command) error)}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commands.register("feeds", handlerFeeds)
	commands.register("follow", middlewareLoggedIn(handlerFollow))
	commands.register("following", middlewareLoggedIn(handlerFollowing))
	commands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	commands.register("browse", middlewareLoggedIn(handlerBrowse))
	return commands
}
