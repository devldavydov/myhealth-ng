package main

import (
	"log"

	"github.com/devldavydov/myhealth-ng/internal/service"
)

func main() {
	config, err := service.ConfigFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	application := service.New(config)
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
