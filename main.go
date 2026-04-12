package main

import (
	"fmt"

	"github.com/Eval-99/aggregator/internal/config"
)

func main() {
	configStruct, err := config.Read()
	if err != nil {
		fmt.Println("Could not read file")
		fmt.Println(err)
		return
	}

	configStruct.SetUser("Some User")

	fmt.Println(configStruct)
}
