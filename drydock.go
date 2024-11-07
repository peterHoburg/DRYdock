package main

import (
	"drydock/internal"
)

func main() {

	dockerFilePaths, e := internal.FindFiles("docker-compose\\.y(?:a)?ml")
	if e != nil {
		panic(e)
	}
	for _, dockerfilePath := range dockerFilePaths {
		internal.LoadComposeFile(dockerfilePath)
	}
	err := internal.WriteComposeFile("docker-compose.yml", []byte("appended some data\n"))

	if err != nil {
		println(err.Error())
	}
}
