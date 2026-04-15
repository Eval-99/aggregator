package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Eval-99/aggregator/internal/config"
	"github.com/Eval-99/aggregator/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	configStruct, err := config.Read()
	if err != nil {
		log.Fatalf("Could not read config file: %v", err)
		return
	}

	db, err := sql.Open("postgres", configStruct.DbUrl)
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}

	dbQueries := database.New(db)

	stateStruct := state{config: &configStruct, db: dbQueries}
	registeredCommands := commands{commandNames: make(map[string]func(*state, command) error)}

	registeredCommands.register("login", handlerLogin)
	registeredCommands.register("register", handlerRegister)
	registeredCommands.register("reset", handlerReset)
	registeredCommands.register("users", handlerUsers)
	registeredCommands.register("agg", handlerAgg)
	registeredCommands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	registeredCommands.register("feeds", handlerFeeds)
	registeredCommands.register("follow", middlewareLoggedIn(handlerFollow))
	registeredCommands.register("following", middlewareLoggedIn(handlerFollowing))
	registeredCommands.register("unfollow", middlewareLoggedIn(handlerUnfollow))

	args := os.Args
	if len(args) < 2 {
		fmt.Println("No command given...")
		os.Exit(1)
	}

	commandToRun := command{args[1], args[2:]}
	err = registeredCommands.run(&stateStruct, commandToRun)
	if err != nil {
		fmt.Println(err)
	}
}
