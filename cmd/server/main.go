package main

import (
	"log"
	"os"

	"github.com/devldavydov/myhealth-ng/internal/service"
)

func main() {
	config, err := service.ConfigFromArgs(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	application, err := service.New(config)
	if err != nil {
		log.Fatal(err)
	}
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
