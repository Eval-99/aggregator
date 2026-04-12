package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Eval-99/aggregator/internal/config"
)

func main() {
	configStruct, err := config.Read()
	if err != nil {
		log.Fatalf("Could not read config file: %v", err)
		return
	}

	stateStruct := state{config: &configStruct}
	registeredCommands := commands{commandNames: make(map[string]func(*state, command) error)}
	registeredCommands.register("login", handlerLogin)

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
