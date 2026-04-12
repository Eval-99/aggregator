package main

import (
	"fmt"
	"log"

	"github.com/Eval-99/aggregator/internal/config"
)

func main() {
	configStruct, err := config.Read()
	if err != nil {
		log.Fatalf("Could not read config file: %v", err)
		return
	}

	err = configStruct.SetUser("Some User")
	if err != nil {
		log.Fatalf("Could not set username: %v", err)
		return
	}

	fmt.Println(configStruct)
}
