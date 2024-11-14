package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

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

func (c Compose) String() string {
	return fmt.Sprintf("Name: %s, Path: %s, Active: %t, Environment: %s, IsRootCompose: %t", c.Name, c.Path, *c.Active, *c.Environment, *c.IsRootCompose)
}
func Pointer[T any](d T) *T {
	return &d
}

// LoadComposeFile loads the docker-compose file into github.com/compose-spec/compose-go/v2/types.Project object.
// Also checks the compose files for basic correctness.
func LoadComposeFile(compose *Compose) (*Compose, error) {
	ctx := context.Background()

	options, err := cli.NewProjectOptions(
		[]string{compose.Path},
		cli.WithoutEnvironmentResolution,
	)
	if err != nil {
		return nil, err
	}

	project, err := options.LoadProject(ctx)
	if err != nil {
		return nil, err
	}
	compose.Project = project
	compose.Name = project.Name

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
func GenerateComposeCommand(compose *Compose) []string {
	composeCommand := []string{"compose", "-f", compose.Path, "up", "--build", "-d"}
	return composeCommand
}

// CombineComposeFiles merges multiple Docker Compose project files into a single project file.
// It iterates over each service in the provided compose files and adds them to the combined compose file.
func CombineComposeFiles(composeFiles []*Compose, combinedComposeFile *Compose) *Compose {
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
func SetProjectName(compose *Compose, projectName string) *Compose {
	compose.Project.Name = projectName
	return compose
}

func GetAllComposeFiles() (*Compose, []*Compose, error) {
	dockerComposeRegex, err := regexp.Compile("docker-compose\\.ya?ml")
	if err != nil {
		return nil, nil, err
	}

	childComposeFilePaths, err := FindFilesRecursively(dockerComposeRegex)
	composeFiles := make([]*Compose, 0)
	for _, path := range childComposeFilePaths {
		composeFiles = append(composeFiles, &Compose{
			Name:          "",
			Path:          path,
			Active:        nil,
			Environment:   nil,
			IsRootCompose: nil,
			Project:       nil,
		})
	}
	rootComposeFile, childComposeFiles, err := LoadAndOrganizeComposeFiles(composeFiles)

	if err != nil {
		return nil, nil, err
	}
	return rootComposeFile, childComposeFiles, nil
}

func LoadAndOrganizeComposeFiles(composeFiles []*Compose) (*Compose, []*Compose, error) {
	updatedComposeFiles := make([]*Compose, 0)
	for _, compose := range composeFiles {
		composeFile, err := LoadComposeFile(compose)
		if err != nil {
			log.Println(compose)
			return nil, nil, err
		}
		updatedComposeFiles = append(updatedComposeFiles, composeFile)
	}
	rootComposeFile, childComposeFiles, err := PickRootComposeFile(updatedComposeFiles)
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

func RunComposeFiles(composeFiles []*Compose) ([]byte, error) {
	projectName := fmt.Sprintf("project-%d", time.Now().Unix())
	networkName := fmt.Sprintf("network-%d", time.Now().Unix())
	envFilePath := "/home/peter/GolandProjects/DRYdock/testdata/example-repo-setup/.example-env-vars" // TODO generate the file path based on env that is being run

	newDockerComposePath, err := filepath.Abs(fmt.Sprintf("docker-compose-%d.yml", time.Now().Unix()))
	if err != nil {
		log.Println(err)
	}

	rootComposeFile, childComposeFiles, err := LoadAndOrganizeComposeFiles(composeFiles)
	if err != nil {
		log.Println(err)
	}
	combinedComposeFile := SetCombinedDepends(childComposeFiles, rootComposeFile)
	combinedComposeFile = CombineComposeFiles(childComposeFiles, combinedComposeFile)
	combinedComposeFile = SetNetwork(combinedComposeFile, networkName)
	combinedComposeFile = SetEnvFile(combinedComposeFile, envFilePath)
	combinedComposeFile = SetProjectName(combinedComposeFile, projectName)
	combinedComposeFile.Path = newDockerComposePath

	combinedComposeFileYaml, err := combinedComposeFile.Project.MarshalYAML()
	if err != nil {
		log.Println(err)
	}

	err = WriteComposeFile(newDockerComposePath, combinedComposeFileYaml)
	if err != nil {
		log.Println(err)
	}

	composeCommand := GenerateComposeCommand(combinedComposeFile)
	cmd := exec.Command("docker", composeCommand...)
	output, err := cmd.CombinedOutput()
	println(string(output))

	if err != nil {
		log.Println(err)
	}
	return output, nil
}
