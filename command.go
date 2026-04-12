package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/Eval-99/aggregator/internal/config"
)

type state struct {
	config *config.Config
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
		error := errors.New("The login command expects a single argument, the username.")
		os.Exit(1)
		return error
	}

	err := s.config.SetUser(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Username has been set to %s\n", cmd.args[0])

	return nil
}
