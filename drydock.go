package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/compose-spec/compose-go/v2/types"

	"drydock/internal"
)

// TODO
// Set env file per service
// Add tests

func initLogger() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}
func main() {
	initLogger()
	projectName := fmt.Sprintf("project-%d", time.Now().Unix())
	networkName := fmt.Sprintf("network-%d", time.Now().Unix())
	envFilePath := "/home/peter/GolandProjects/DRYdock/testdata/example-repo-setup/.example-env-vars" // TODO generate the file path based on env that is being run

	newDockerComposePath, err := filepath.Abs(fmt.Sprintf("docker-compose-%d.yml", time.Now().Unix()))
	if err != nil {
		log.Fatal(err)
	}

	dockerComposeRegex, err := regexp.Compile("docker-compose\\.y(?:a)?ml")
	if err != nil {
		log.Fatal(err)
	}

	childComposeFilePaths, err := internal.FindFilesInChildDirs(dockerComposeRegex)
	if err != nil {
		log.Fatal(err)
	}

	rootComposePath, err := internal.FindFileInCurrentDir(dockerComposeRegex)
	if err != nil {
		log.Fatal(err)
	}

	rootComposeFile, err := internal.LoadComposeFile(rootComposePath, projectName)
	if err != nil {
		log.Println(rootComposePath)
		log.Fatal(err)
	}

	childComposeFiles := make([]*types.Project, 0)
	for _, composeFilePath := range childComposeFilePaths {
		composeFile, err := internal.LoadComposeFile(composeFilePath, projectName)
		if err != nil {
			log.Println(composeFilePath)
			log.Fatal(err)
		}
		childComposeFiles = append(childComposeFiles, composeFile)
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
