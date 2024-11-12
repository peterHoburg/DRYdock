package internal

import (
	"context"
	"errors"
	"log"
	"os"
	"regexp"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
)

// LoadComposeFile loads the docker-compose file into github.com/compose-spec/compose-go/v2/types.Project object.
// Also checks the compose files for basic correctness.
func LoadComposeFile(composePath string, projectName string) (*types.Project, error) {
	ctx := context.Background()

	options, err := cli.NewProjectOptions(
		[]string{composePath},
		cli.WithoutEnvironmentResolution,
		cli.WithName(projectName),
	)
	if err != nil {
		return nil, err
	}

	project, err := options.LoadProject(ctx)
	if err != nil {
		return nil, err
	}

	err = CheckComposeFile(project)
	if err != nil {
		return nil, err
	}

	return project, nil

}

// WriteComposeFile writes the given data to a file specified by composePath, creating it if it doesn't exist.
func WriteComposeFile(composePath string, data []byte) error {
	f, err := os.OpenFile(composePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close() // ignore error; Write error takes precedence
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

// GenerateComposeCommand creates a Docker Compose command using the specified compose file path with '-f'.
func GenerateComposeCommand(composePath string) []string {
	composeCommand := []string{"compose", "-f", composePath, "up", "--build", "-d"}
	return composeCommand
}

// CombineComposeFiles merges multiple Docker Compose project files into a single project file.
// It iterates over each service in the provided compose files and adds them to the combined compose file.
func CombineComposeFiles(composeFiles []*types.Project, combinedComposeFile *types.Project) *types.Project {
	for _, c := range composeFiles {
		for k, v := range c.Services {
			combinedComposeFile.Services[k] = v
		}
	}
	return combinedComposeFile
}

func SetCombinedDepends(composeFiles []*types.Project, combinedCompose *types.Project) *types.Project {
	service := types.ServiceConfig{
		Name:        "combined",
		Build:       nil,
		Command:     nil,
		DependsOn:   nil,
		Entrypoint:  nil,
		Environment: nil,
		EnvFiles:    nil,
		Ports:       nil,
	}

	if combinedCompose != nil {
		service = combinedCompose.Services["combined"]
	}

	dependsOn := map[string]types.ServiceDependency{}
	for _, composeFile := range composeFiles {
		for _, service := range composeFile.Services {
			dependsOn[service.Name] = types.ServiceDependency{Required: true, Condition: "service_started"}

		}
	}
	service.DependsOn = dependsOn
	combinedCompose.Services["combined"] = service
	return combinedCompose
}

func CheckComposeFile(composeFile *types.Project) error {
	for _, service := range composeFile.Services {
		build := service.Build
		if build == nil {
			return nil
		}
		if build.Dockerfile == "." {
			return errors.New("int docker-compose build.dockerfile cannot be '.' try './Dockerfile'")
		}
	}
	return nil
}

func SetNetwork(combinedCompose *types.Project, networkName string) *types.Project {
	if combinedCompose.Networks == nil {
		combinedCompose.Networks = map[string]types.NetworkConfig{}
	}
	delete(combinedCompose.Networks, "default")
	combinedCompose.Networks[networkName] = types.NetworkConfig{
		Driver: "bridge",
	}
	for _, service := range combinedCompose.Services {
		delete(service.Networks, "default")
		service.Networks[networkName] = &types.ServiceNetworkConfig{}
	}
	return combinedCompose
}

func SetEnvFile(combinedCompose *types.Project, envFilePath string) *types.Project {
	for serviceName, service := range combinedCompose.Services {
		service.EnvFiles = []types.EnvFile{
			{
				Path:     envFilePath,
				Required: true,
				Format:   "",
			},
		}
		combinedCompose.Services[serviceName] = service
	}
	return combinedCompose
}

func GetAllComposeFiles(projectName string) (*types.Project, []*types.Project, error) {
	dockerComposeRegex, err := regexp.Compile("docker-compose\\.y(?:a)?ml")
	if err != nil {
		return nil, nil, err
	}

	childComposeFilePaths, err := FindFilesRecursively(dockerComposeRegex)
	if err != nil {
		return nil, nil, err
	}

	composeFiles := make([]*types.Project, 0)
	for _, composeFilePath := range childComposeFilePaths {
		composeFile, err := LoadComposeFile(composeFilePath, projectName)
		if err != nil {
			log.Println(composeFilePath)
			return nil, nil, err
		}
		composeFiles = append(composeFiles, composeFile)
	}
	rootComposeFile, childComposeFiles, err := PickRootComposeFile(composeFiles)
	if err != nil {
		return nil, nil, err
	}
	return rootComposeFile, childComposeFiles, nil
}

func PickRootComposeFile(composeFiles []*types.Project) (*types.Project, []*types.Project, error) {
	var rootCompose *types.Project
	for i, composeFile := range composeFiles {
		for _, service := range composeFile.Services {
			if service.Name == "combined" {
				if rootCompose != nil {
					return nil, nil, errors.New("multiple root compose files found, root compose needs a service called combined")
				}
				rootCompose = composeFile
				composeFiles = append(composeFiles[:i], composeFiles[i+1:]...)
			}
		}
	}

	if rootCompose == nil {
		return nil, nil, errors.New("no root compose file found, root compose needs a service called combined")
	}
	return rootCompose, composeFiles, nil
}
