package main

import (
	"drydock/api"
	"drydock/config"
)

// TODO
// Set env file per service
// Add tests

func main() {
	config.LoadConfig()
	config.InitLogger()
	api.Start()
}
