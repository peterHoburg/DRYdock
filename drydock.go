package main

import (
	"drydock/api"
)

// TODO
// Set env file per service
// Add tests

func main() {
	InitLogger()
	api.Start()
}
