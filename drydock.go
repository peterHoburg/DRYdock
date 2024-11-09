package main

import (
	"log"
	"os/exec"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/types"

	"drydock/internal"
)

// TODO
// Generate a specific name for the docker-compose file
// add custom network,
// change the name of the project name:
func initLogger() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}
func main() {
	initLogger()

	composeFilePaths, e := internal.FindFiles("docker-compose\\.y(?:a)?ml")
	if e != nil {
		log.Fatal(e)
	}

	composeFiles := make([]*types.Project, 0)
	for _, composeFilePath := range composeFilePaths { // We are assuming the "combined" docker compose file will be first in the list of docker file paths. This might be a mistake...
		composeFile, err := internal.LoadComposeFile(composeFilePath)
		if err != nil {
			log.Fatal(err)
		}

		err = internal.CheckComposeFile(composeFile)
		if err != nil {
			log.Println(composeFilePath)
			log.Fatal(err)
		}
		composeFiles = append(composeFiles, composeFile)
	}
	// TODO this will break if composeFiles len < 2
	childComposeFiles := composeFiles[1:]
	combinedComposeFile := composeFiles[0]
	combinedComposeFile, err := internal.SetCombinedDepends(childComposeFiles, combinedComposeFile)
	if err != nil {
		log.Fatal(err)
	}
	combinedComposeFile, err = internal.CombineComposeFiles(childComposeFiles, combinedComposeFile)
	if err != nil {
		log.Fatal(err)
	}
	combinedComposeFile, err = internal.SetNetwork(combinedComposeFile)
	if err != nil {
		log.Fatal(err)
	}

	// TODO generate the file path based on env that is being run
	combinedComposeFile, err = internal.SetEnvironmentFile(combinedComposeFile, "/home/peter/GolandProjects/DRYdock/testdata/example-repo-setup/.example-env-vars")
	if err != nil {
		log.Fatal(err)
	}

	newDockerComposePath, err := filepath.Abs("./docker-compose-new.yml")
	if err != nil {
		log.Fatal(err)
	}
	combinedComposeFileYaml, err := combinedComposeFile.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}
	err = internal.WriteComposeFile(newDockerComposePath, combinedComposeFileYaml)

	if err != nil {
		log.Fatal(err)
	}

	composeCommand, err := internal.GenerateComposeCommand(newDockerComposePath)
	cmd := exec.Command("docker", composeCommand...)
	output, err := cmd.CombinedOutput()
	println(string(output))

	if err != nil {
		log.Fatal(err)
	}

}
