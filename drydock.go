package main

import (
	"drydock/api"
	"github.com/rs/zerolog/log"
)

// TODO
// Set env file per service
// Add tests

func main() {
	err := LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config")
	}
	InitLogger()
	api.Start()
}
