package internal

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/rs/zerolog/log"
)

type Compose struct {
	Name             string
	Path             string
	Active           bool
	Environment      string
	IsRootCompose    bool
	Project          *types.Project
	EnvFilePath      string
	EnvVarFileFormat string
}

func (c Compose) String() string {
	return fmt.Sprintf("Name: %s, Path: %s, Active: %t, Environment: %s, IsRootCompose: %t", c.Name, c.Path, c.Active, c.Environment, c.IsRootCompose)
}

type ComposeRunData struct {
	ComposeFiles                   []*Compose
	ProjectName                    string
	NetworkName                    string
	NewDockerComposePath           string
	RemoveOrphans                  bool
	AlwaysRecreateDeps             bool
	ComposeCommand                 string
	StopAllContainersBeforeRunning bool
	PreRunCommand                  string
	EnvVarFileSetupCommand         string
}

func (c ComposeRunData) LoadFromForm(form url.Values) (ComposeRunData, error) {
	var defaultEnvironmentSelect string
	var environment string
	var envVarFileFormat string

	for k, v := range form {
		if k == "defaultEnvironmentSelect" {
			defaultEnvironmentSelect = v[0]
			continue
		}
		if k == "removeOrphans" {
			c.RemoveOrphans = true
			continue
		}
		if k == "alwaysRecreateDeps" {
			c.AlwaysRecreateDeps = true
			continue
		}
		if k == "stopAllContainersBeforeRunning" {
			c.StopAllContainersBeforeRunning = true
			continue
		}
		if k == "composeCommand" {
			c.ComposeCommand = v[0]
			continue
		}
		if k == "composeFileName" {
			if v[0] == "" {
				continue
			}
			newDockerComposePath, err := filepath.Abs(strings.Replace(v[0], "[[TIMESTAMP]]", fmt.Sprintf("%d", time.Now().Unix()), 1))

			if err != nil {
				return c, err
			}
			c.NewDockerComposePath = newDockerComposePath
			continue
		}

		if k == "preRunCommand" {
			c.PreRunCommand = v[0]
			continue
		}
		if k == "envVarFileFormat" {
			envVarFileFormat = v[0]
			continue
		}
		if k == "envVarFileSetupCommand" {
			c.EnvVarFileSetupCommand = v[0]
			continue
		}
	}

	for k, v := range form {
		if (len(v) > 1 && v[1] == "on") || k == "rootComposeFile" {
			if v[0] == "default" || k == "rootComposeFile" {
				environment = defaultEnvironmentSelect
			} else {
				environment = v[0]
			}
			if k == "rootComposeFile" {
				c.ComposeFiles = append(c.ComposeFiles, &Compose{
					Path:             v[0] + "/docker-compose.yml",
					Active:           true,
					Environment:      environment,
					EnvVarFileFormat: envVarFileFormat,
				})
				continue
			}
			c.ComposeFiles = append(c.ComposeFiles, &Compose{
				Path:             k + "/docker-compose.yml",
				Active:           true,
				Environment:      environment,
				EnvVarFileFormat: envVarFileFormat,
			})
		}
	}
	return c, nil
}

func (c ComposeRunData) ReplacePlaceholders() ComposeRunData {
	c.NewDockerComposePath = strings.Replace(c.NewDockerComposePath, "[[TIMESTAMP]]", fmt.Sprintf("%d", time.Now().Unix()), 1)
	c.ComposeCommand = strings.Replace(c.ComposeCommand, "[[COMPOSE_FILE_NAME]]", c.NewDockerComposePath, 1)
	return c
}

type ComposeRunDataReturn struct {
	ComposeFile *Compose
	Command     []string
}

func Pointer[T any](d T) *T {
	return &d
}

// LoadComposeFile loads the docker-compose file into github.com/compose-spec/compose-go/v2/types.Project object.
// Also checks the compose files for basic correctness.
func LoadComposeFile(compose *Compose) (*Compose, error) {
	log.Debug().Msg(fmt.Sprintf("Loading compose file: %s", compose.Path))
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

	err = checkComposeFile(compose)
	if err != nil {
		return nil, err
	}

	return compose, nil

}

// WriteComposeFile writes the given data to a file specified by composePath, creating it if it doesn't exist.
func WriteComposeFile(composePath string, data []byte) error {
	log.Debug().Msg(fmt.Sprintf("Writing compose file: %s", composePath))
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
func GenerateComposeCommand(compose *Compose, composeRunData ComposeRunData) []string {
	log.Debug().Msg(fmt.Sprintf("Generating compose command for compose file: %s", compose.Path))
	composeCommand := []string{"compose", "-f", compose.Path}
	if composeRunData.ComposeCommand != "" {
		composeCommand = strings.Fields(composeRunData.ComposeCommand)
	} else {
		composeCommand = append(composeCommand, "up", "--build", "-d")
	}

	if composeRunData.RemoveOrphans {
		composeCommand = append(composeCommand, "--remove-orphans")
	}
	if composeRunData.AlwaysRecreateDeps {
		composeCommand = append(composeCommand, "--always-recreate-deps")
	}
	log.Info().Msg(strings.Join(composeCommand, " "))
	return composeCommand
}

// CombineComposeFiles merges multiple Docker Compose project files into a single project file.
// It iterates over each service in the provided compose files and adds them to the combined compose file.
func CombineComposeFiles(composeFiles []*Compose, combinedComposeFile *Compose) *Compose {
	log.Debug().Msg("Combining compose files")
	for _, c := range composeFiles {
		for k, v := range c.Project.Services {
			combinedComposeFile.Project.Services[k] = v
		}
	}
	return combinedComposeFile
}

func setCombinedDepends(composeFiles []*Compose, combinedCompose *Compose) *Compose {
	log.Debug().Msg("Setting combined service depends on all services")
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

func checkComposeFile(composeFile *Compose) error {
	log.Debug().Msg(fmt.Sprintf("Checking compose file: %s", composeFile.Path))
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

func setNetwork(combinedCompose *Compose, networkName string) *Compose {
	log.Debug().Msg(fmt.Sprintf("Setting network: %s in compose: %s", networkName, combinedCompose.Path))
	if combinedCompose.Project.Networks == nil {
		log.Trace().Msg("Networks is nil, creating new map")
		combinedCompose.Project.Networks = map[string]types.NetworkConfig{}
	}
	delete(combinedCompose.Project.Networks, "default")
	combinedCompose.Project.Networks[networkName] = types.NetworkConfig{
		Driver: "bridge",
	}
	for _, service := range combinedCompose.Project.Services {
		log.Trace().Msg(fmt.Sprintf("Setting network: %s for service: %s", networkName, service.Name))
		delete(service.Networks, "default")
		service.Networks[networkName] = &types.ServiceNetworkConfig{}
	}
	return combinedCompose
}

func setEnvFile(compose *Compose) *Compose {
	log.Debug().Msg(fmt.Sprintf("Setting env file: %s in compose: %s", compose.EnvFilePath, compose.Path))
	envPath, err := filepath.Abs(strings.Replace(compose.EnvVarFileFormat, "[[ENVIRONMENT]]", compose.Environment, 1))
	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("Error setting env file: %s in compose: %s", envPath, compose.Path))
		return compose
	}
	for serviceName, service := range compose.Project.Services {
		service.EnvFiles = []types.EnvFile{
			{
				Path:     envPath,
				Required: true,
				Format:   "",
			},
		}
		compose.Project.Services[serviceName] = service
	}
	return compose
}
func setProjectName(compose *Compose, projectName string) *Compose {
	log.Debug().Msg(fmt.Sprintf("Setting project name: %s in compose: %s", projectName, compose.Path))
	compose.Project.Name = projectName
	return compose
}

func GetAllComposeFiles() (*Compose, []*Compose, error) {
	log.Debug().Msg("Getting all compose files")
	dockerComposeRegex, err := regexp.Compile("docker-compose\\.ya?ml")
	if err != nil {
		return nil, nil, err
	}

	childComposeFilePaths, err := FindFilesRecursively(dockerComposeRegex)
	composeFiles := make([]*Compose, 0)
	for _, path := range childComposeFilePaths {
		log.Trace().Msg(fmt.Sprintf("Found compose file: %s", path))
		composeFiles = append(composeFiles, &Compose{
			Name:          "",
			Path:          path,
			Active:        false,
			Environment:   "",
			IsRootCompose: false,
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
	log.Debug().Msg("Loading and organizing compose files")
	updatedComposeFiles := make([]*Compose, 0)
	for _, compose := range composeFiles {
		log.Trace().Msg(fmt.Sprintf("Loading compose file: %s", compose.Path))
		composeFile, err := LoadComposeFile(compose)
		if err != nil {
			log.Error().Err(err).Msg(compose.String())
			return nil, nil, err
		}
		composeFile = setEnvFile(composeFile)
		updatedComposeFiles = append(updatedComposeFiles, composeFile)
	}

	rootComposeFile, childComposeFiles, err := PickRootComposeFile(updatedComposeFiles)
	return rootComposeFile, childComposeFiles, err
}

func PickRootComposeFile(composeFiles []*Compose) (*Compose, []*Compose, error) {
	log.Debug().Msg("Picking root compose file")
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

func stopAllContainers() error {
	log.Debug().Msg("Stopping all containers")

	allRunningContainers, err := exec.Command("docker", "ps", "-q").CombinedOutput()
	if len(allRunningContainers) == 0 {
		log.Info().Msg("No running containers to stop")
		return nil
	}
	strippedAllRunningContainers := strings.TrimSpace(string(allRunningContainers))
	splitContainers := strings.Split(strippedAllRunningContainers, "\n")

	command := []string{"docker", "stop"}
	command = append(command, splitContainers...)
	log.Info().Msg(fmt.Sprintf("Running command: %s", strings.Join(command, " ")))

	cmd := exec.Command(command[0], command[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Msg(string(output))
		return err
	}
	return nil
}

func runEnvVarSetupCommand(composeRunData ComposeRunData) error {
	if composeRunData.EnvVarFileSetupCommand == "" {
		return nil
	}
	for _, compose := range composeRunData.ComposeFiles {
		command := strings.Replace(composeRunData.EnvVarFileSetupCommand, "[[ENVIRONMENT]]", compose.Environment, 1)
		log.Debug().Msg(fmt.Sprintf("Running env var setup command: %s", command))
		cmd := exec.Command("sh", "-c", command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("Error running env var setup command: %s \nOutput: %s", command, string(output)))
			return err
		}
		log.Info().Msg(string(output))
	}
	return nil
}

func ComposeFilesToRunCommand(composeRunData ComposeRunData) (*ComposeRunDataReturn, error) {
	log.Debug().Msg("Converting compose files to run command")
	if composeRunData.StopAllContainersBeforeRunning {
		err := stopAllContainers()
		if err != nil {
			return nil, err
		}
	}
	rootComposeFile, childComposeFiles, err := LoadAndOrganizeComposeFiles(composeRunData.ComposeFiles)
	if err != nil {
		return nil, err
	}
	err = runEnvVarSetupCommand(composeRunData)
	if err != nil {
		return nil, err
	}
	combinedComposeFile := setCombinedDepends(childComposeFiles, rootComposeFile)
	combinedComposeFile = CombineComposeFiles(childComposeFiles, combinedComposeFile)
	combinedComposeFile = setNetwork(combinedComposeFile, composeRunData.NetworkName)
	combinedComposeFile = setProjectName(combinedComposeFile, composeRunData.ProjectName)
	combinedComposeFile.Path = composeRunData.NewDockerComposePath

	composeCommand := GenerateComposeCommand(combinedComposeFile, composeRunData)
	return &ComposeRunDataReturn{
		ComposeFile: combinedComposeFile,
		Command:     composeCommand,
	}, nil
}
