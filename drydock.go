package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"time"

	"drydock/api"
	"drydock/internal"
)

// TODO
// Set env file per service
// Add tests

func initLogger() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}
func main() {
	api.Start()
}
func mainCli() {
	initLogger()
	projectName := fmt.Sprintf("project-%d", time.Now().Unix())
	networkName := fmt.Sprintf("network-%d", time.Now().Unix())
	envFilePath := "/home/peter/GolandProjects/DRYdock/testdata/example-repo-setup/.example-env-vars" // TODO generate the file path based on env that is being run

	newDockerComposePath, err := filepath.Abs(fmt.Sprintf("docker-compose-%d.yml", time.Now().Unix()))
	if err != nil {
		log.Fatal(err)
	}

	rootComposeFile, childComposeFiles, err := internal.GetAllComposeFiles(projectName)
	if err != nil {
		log.Fatal(err)
	}
	combinedComposeFile := internal.SetCombinedDepends(childComposeFiles, rootComposeFile)
	combinedComposeFile = internal.CombineComposeFiles(childComposeFiles, combinedComposeFile)
	combinedComposeFile = internal.SetNetwork(combinedComposeFile, networkName)
	combinedComposeFile = internal.SetEnvFile(combinedComposeFile, envFilePath)

	combinedComposeFileYaml, err := combinedComposeFile.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}

	err = internal.WriteComposeFile(newDockerComposePath, combinedComposeFileYaml)
	if err != nil {
		log.Fatal(err)
	}

	composeCommand := internal.GenerateComposeCommand(newDockerComposePath)
	cmd := exec.Command("docker", composeCommand...)
	output, err := cmd.CombinedOutput()
	println(string(output))

	if err != nil {
		log.Fatal(err)
	}

}
