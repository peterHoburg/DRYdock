package main

import (
	"drydock/internal"
	"github.com/compose-spec/compose-go/v2/cli"
)

func main() {
	dockerFilePaths, e := internal.FindFiles("Dockerfile")
	if e != nil {
		panic(e)
	}

}
