package main

import (
	"log"
	"os/exec"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/types"

	"drydock/internal"
)

func initLogger() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}
func main() {

	dockerFilePaths, e := internal.FindFiles("docker-compose\\.y(?:a)?ml")
	if e != nil {
		panic(e)
	}

	composeFiles := make([]*types.Project, 0)
	for _, dockerfilePath := range dockerFilePaths {
		composeFile, err := internal.LoadComposeFile(dockerfilePath)
		if err != nil {
			log.Fatal(err)
		}

		err = internal.CheckComposeFile(composeFile)
		if err != nil {
			log.Println(dockerfilePath)
			log.Fatal(err)
		}
		composeFiles = append(composeFiles, composeFile)
	}
	combinedComposeFile, err := internal.CombineComposeFiles(composeFiles)
	newDockerComposePath, err := filepath.Abs("./docker-compose.yml")
	if err != nil {
		panic(err)
	}
	combinedComposeFileYaml, err := combinedComposeFile.MarshalYAML()
	if err != nil {
		println(err.Error())
	}
	err = internal.WriteComposeFile(newDockerComposePath, combinedComposeFileYaml)

	if err != nil {
		println(err.Error())
	}

	composeCommand, err := internal.GenerateComposeCommand(newDockerComposePath)
	cmd := exec.Command("docker", composeCommand...)
	output, err := cmd.CombinedOutput()
	println(string(output))

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

}
