package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
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
	if len(cmd.args) == 0 || len(cmd.args) > 1 {
		fmt.Println("The agg command expects a single argument, the durantion.")
		os.Exit(1)
	}

	time_parsed, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		fmt.Println("Time string incorrectly formatted.")
		return err
	}

	time_between_reqs := time.Duration(time_parsed)

	fmt.Printf("Collecting feeds every %s\n", cmd.args[0])
	ctx := context.Background()
	ticker := time.NewTicker(time_between_reqs)
	for ; ; <-ticker.C {
		scrapeFeeds(s, ctx)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 || len(cmd.args) < 2 || len(cmd.args) > 2 {
		fmt.Println("The addfeed command expects two arguments, the name and url.")
		os.Exit(1)
	}

	ctx := context.Background()

	query := database.AddFeedParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.args[0], Url: cmd.args[1], UserID: user.ID}

	feed, err := s.db.AddFeed(ctx, query)
	if err != nil {
		return err
	}

	err = handlerFollow(s, cmd, user)
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

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		fmt.Println("The follow command expects a single argument, the feed URL.")
		os.Exit(1)
	}

	var url string
	if cmd.name == "addfeed" {
		url = cmd.args[1]
	} else {
		url = cmd.args[0]
	}

	ctx := context.Background()

	feed, err := s.db.GetFeed(ctx, url)
	if err != nil {
		fmt.Printf("The feed %s does not exist\n", url)
		os.Exit(1)
	}

	query := database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), UserID: user.ID, FeedID: feed.ID}

	feedRow, err := s.db.CreateFeedFollow(ctx, query)
	if err != nil {
		return err
	}

	fmt.Printf("User %v is now following '%v'\n", feedRow.UserName, feedRow.FeedName)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 0 {
		fmt.Println("The following command expects no argument.")
		os.Exit(1)
	}

	ctx := context.Background()

	feedsFollowing, err := s.db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, follow := range feedsFollowing {
		fmt.Printf(" - '%v'\n", follow.FeedName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		fmt.Println("The unfollow command expects a single argument, the feed URL.")
		os.Exit(1)
	}

	ctx := context.Background()

	feed, err := s.db.GetFeed(ctx, cmd.args[0])
	if err != nil {
		fmt.Printf("The feed %s does not exist\n", cmd.args[0])
		os.Exit(1)
	}

	query := database.DeleteFollowFeedParams{UserID: user.ID, FeedID: feed.ID}

	err = s.db.DeleteFollowFeed(ctx, query)
	if err != nil {
		return err
	}

	fmt.Printf("Unfollowed feed '%v'\n", feed.Name)

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	if len(cmd.args) > 1 {
		fmt.Println("The browse command expects zero or a single argument, the amount of posts.")
		os.Exit(1)
	}

	var amount int32
	if len(cmd.args) == 1 {
		res, err := strconv.Atoi(cmd.args[0])
		amount = int32(res)
		if err != nil {
			fmt.Println("Argument not a number")
			os.Exit(1)
		}
	} else if len(cmd.args) == 0 {
		amount = 2
	}

	ctx := context.Background()
	query := database.GetPostsForUserParams{UserID: user.ID, Limit: amount}
	posts, err := s.db.GetPostsForUser(ctx, query)
	if err != nil {
		fmt.Println("Could not get user posts")
		os.Exit(1)
	}

	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return nil
}

func scrapeFeeds(s *state, ctx context.Context) error {
	feed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return err
	}

	err = s.db.MarkFeedFetched(ctx, database.MarkFeedFetchedParams{ID: feed.ID, UpdatedAt: time.Now()})
	if err != nil {
		return err
	}

	res, err := fetchFeed(ctx, feed.Url)

	fmt.Printf("'%v' has been updated\n", res.Channel.Title)
	for _, item := range res.Channel.Item {
		pubTime, _ := time.Parse("Sun, 08 Jan 2023 00:00:00 +0000", item.PubDate)
		query := database.CreatePostParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Title: item.Title, Url: item.Link, Description: item.Description, PublishedAt: pubTime, FeedID: feed.ID}
		s.db.CreatePost(ctx, query)
	}

	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		ctx := context.Background()
		user, err := s.db.GetUser(ctx, s.config.CurrentUserName)
		if err != nil {
			fmt.Printf("The user %s does not exist\n", s.config.CurrentUserName)
			os.Exit(1)
		}
		return handler(s, cmd, user)
	}
}
