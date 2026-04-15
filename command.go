package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Eval-99/aggregator/internal/config"
	"github.com/Eval-99/aggregator/internal/database"
	"github.com/google/uuid"
)

type state struct {
	config *config.Config
	db     *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	commandNames map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	commandFunc, ok := c.commandNames[cmd.name]
	if !ok {
		return fmt.Errorf("%s is not a registered command", cmd.name)
	}

	return commandFunc(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandNames[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 || len(cmd.args) > 1 {
		fmt.Println("The login command expects a single argument, the username.")
		os.Exit(1)
	}

	ctx := context.Background()

	_, err := s.db.GetUser(ctx, cmd.args[0])
	if err != nil {
		fmt.Printf("The user %s does not exist\n", cmd.args[0])
		os.Exit(1)
	}

	err = s.config.SetUser(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Username has been set to %s\n", cmd.args[0])

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 || len(cmd.args) > 1 {
		fmt.Println("The register command expects a single argument, the username.")
		os.Exit(1)
	}

	ctx := context.Background()

	_, err := s.db.GetUser(ctx, cmd.args[0])
	if err == nil {
		fmt.Printf("The user %s is already registered\n", cmd.args[0])
		os.Exit(1)
	}

	query := database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.args[0]}

	user, err := s.db.CreateUser(ctx, query)
	if err != nil {
		return err
	}

	err = s.config.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("The user %s was created\n", user.Name)
	fmt.Println(user)

	return nil
}

func handlerReset(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		fmt.Println("The reset command expects no argument.")
		os.Exit(1)
	}

	ctx := context.Background()

	err := s.db.ResetDB(ctx)
	if err != nil {
		fmt.Printf("Error deleting all users: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Success deleting all users and feeds")

	return nil
}

func handlerUsers(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		fmt.Println("The users command expects no argument.")
		os.Exit(1)
	}

	ctx := context.Background()

	users, err := s.db.GetUsers(ctx)
	if err != nil {
		fmt.Printf("Error fetching all users: %v\n", err)
		os.Exit(1)
	}

	currentUser := s.config.CurrentUserName

	for _, user := range users {
		if user.Name == currentUser {
			fmt.Println("* " + user.Name + " (current)")
			continue
		}
		fmt.Println("* " + user.Name)
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		fmt.Println("The agg command expects no argument.")
		os.Exit(1)
	}

	ctx := context.Background()

	feed, err := fetchFeed(ctx, "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}

	fmt.Println(feed)

	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) == 0 || len(cmd.args) < 2 || len(cmd.args) > 2 {
		fmt.Println("The addfeed command expects two arguments, the name and url.")
		os.Exit(1)
	}

	ctx := context.Background()

	user, _ := s.db.GetUser(ctx, s.config.CurrentUserName)

	query := database.AddFeedParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.args[0], Url: cmd.args[1], UserID: user.ID}

	feed, err := s.db.AddFeed(ctx, query)
	if err != nil {
		return err
	}

	fmt.Println(feed)

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		fmt.Println("The feeds command expects no argument.")
		os.Exit(1)
	}

	ctx := context.Background()

	feeds, err := s.db.AllFeeds(ctx)
	if err != nil {
		fmt.Printf("Error fetching all feeds: %v\n", err)
		os.Exit(1)
	}

	for _, feed := range feeds {
		fmt.Println(feed.Name)
		fmt.Printf(" - URL: %v\n", feed.Url)
		fmt.Printf(" - User: %v\n", feed.Name_2)
	}

	return nil
}
