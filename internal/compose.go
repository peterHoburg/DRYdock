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

type Compose struct {
	Name          string
	Path          string
	Active        *bool
	Environment   *string
	IsRootCompose *bool
	Project       *types.Project
}

func Pointer[T any](d T) *T {
	return &d
}

// LoadComposeFile loads the docker-compose file into github.com/compose-spec/compose-go/v2/types.Project object.
// Also checks the compose files for basic correctness.
func LoadComposeFile(composePath string) (*Compose, error) {
	ctx := context.Background()

	options, err := cli.NewProjectOptions(
		[]string{composePath},
		cli.WithoutEnvironmentResolution,
		//cli.WithName(projectName),
	)
	if err != nil {
		return nil, err
	}

	project, err := options.LoadProject(ctx)
	if err != nil {
		return nil, err
	}
	compose := &Compose{
		Name:        project.Name,
		Path:        composePath,
		Active:      nil,
		Environment: nil,
		Project:     project,
	}
	err = CheckComposeFile(compose)
	if err != nil {
		return nil, err
	}

	return compose, nil

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
func GenerateComposeCommand(compose Compose) []string {
	composeCommand := []string{"compose", "-f", compose.Path, "up", "--build", "-d"}
	return composeCommand
}

// CombineComposeFiles merges multiple Docker Compose project files into a single project file.
// It iterates over each service in the provided compose files and adds them to the combined compose file.
func CombineComposeFiles(composeFiles []*Compose, combinedComposeFile Compose) Compose {
	for _, c := range composeFiles {
		for k, v := range c.Project.Services {
			combinedComposeFile.Project.Services[k] = v
		}
	}
	return combinedComposeFile
}

func SetCombinedDepends(composeFiles []*Compose, combinedCompose *Compose) *Compose {
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
		service = combinedCompose.Project.Services["combined"]
	}

	dependsOn := map[string]types.ServiceDependency{}
	for _, composeFile := range composeFiles {
		for _, service := range composeFile.Project.Services {
			dependsOn[service.Name] = types.ServiceDependency{Required: true, Condition: "service_started"}

		}
	}
	service.DependsOn = dependsOn
	combinedCompose.Project.Services["combined"] = service
	return combinedCompose
}

func CheckComposeFile(composeFile *Compose) error {
	for _, service := range composeFile.Project.Services {
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

func SetNetwork(combinedCompose *Compose, networkName string) *Compose {
	if combinedCompose.Project.Networks == nil {
		combinedCompose.Project.Networks = map[string]types.NetworkConfig{}
	}
	delete(combinedCompose.Project.Networks, "default")
	combinedCompose.Project.Networks[networkName] = types.NetworkConfig{
		Driver: "bridge",
	}
	for _, service := range combinedCompose.Project.Services {
		delete(service.Networks, "default")
		service.Networks[networkName] = &types.ServiceNetworkConfig{}
	}
	return combinedCompose
}

func SetEnvFile(combinedCompose *Compose, envFilePath string) *Compose {
	for serviceName, service := range combinedCompose.Project.Services {
		service.EnvFiles = []types.EnvFile{
			{
				Path:     envFilePath,
				Required: true,
				Format:   "",
			},
		}
		combinedCompose.Project.Services[serviceName] = service
	}
	return combinedCompose
}

func GetAllComposeFiles(projectName string) (*Compose, []*Compose, error) {
	dockerComposeRegex, err := regexp.Compile("docker-compose\\.y(?:a)?ml")
	if err != nil {
		return nil, nil, err
	}

	childComposeFilePaths, err := FindFilesRecursively(dockerComposeRegex)
	rootComposeFile, childComposeFiles, err := LoadAndOrganizeComposeFiles(childComposeFilePaths)

	if err != nil {
		return nil, nil, err
	}
	return rootComposeFile, childComposeFiles, nil
}

func LoadAndOrganizeComposeFiles(childComposeFilePaths []string) (*Compose, []*Compose, error) {
	composeFiles := make([]*Compose, 0)
	for _, composeFilePath := range childComposeFilePaths {
		composeFile, err := LoadComposeFile(composeFilePath)
		if err != nil {
			log.Println(composeFilePath)
			return nil, nil, err
		}
		composeFiles = append(composeFiles, composeFile)
	}
	rootComposeFile, childComposeFiles, err := PickRootComposeFile(composeFiles)
	return rootComposeFile, childComposeFiles, err
}
func PickRootComposeFile(composeFiles []*Compose) (*Compose, []*Compose, error) {
	var rootCompose *Compose
	for i, composeFile := range composeFiles {
		for _, service := range composeFile.Project.Services {
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
