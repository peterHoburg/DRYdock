package internal

import (
	"context"
	"errors"
	"os"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
)

func LoadComposeFile(composePath string) (*types.Project, error) {
	projectName := "my_project"
	ctx := context.Background()

	options, err := cli.NewProjectOptions(
		[]string{composePath},
		cli.WithOsEnv,
		cli.WithDotEnv,
		cli.WithName(projectName),
	)
	if err != nil {
		return nil, err
	}

	project, err := options.LoadProject(ctx)
	if err != nil {
		return nil, err
	}

	//// Use the MarshalYAML method to get YAML representation
	//projectYAML, err := project.MarshalYAML()
	//if err != nil {
	//	return nil, err
	//}

	return project, nil

}

func WriteComposeFile(composePath string, data []byte) error {
	f, err := os.OpenFile(composePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close() // ignore error; Write error takes precedence
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

func GenerateComposeCommand(composePath string) ([]string, error) {
	composeCommand := []string{"compose", "-f", composePath, "up", "--build", "-d"}
	return composeCommand, nil
}

func CombineComposeFiles(composeFiles []*types.Project, combinedComposeFile *types.Project) (*types.Project, error) {
	for _, c := range composeFiles {
		for k, v := range c.Services {
			combinedComposeFile.Services[k] = v
		}
	}
	return combinedComposeFile, nil
}

func SetCombinedDepends(composeFiles []*types.Project, combinedCompose *types.Project) (*types.Project, error) {
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
	return combinedCompose, nil
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

func SetNetwork(combinedCompose *types.Project) (*types.Project, error) {
	if combinedCompose.Networks == nil {
		combinedCompose.Networks = map[string]types.NetworkConfig{}
	}
	networkName := "generate-network-name" // TODO generate this
	delete(combinedCompose.Networks, "default")
	combinedCompose.Networks[networkName] = types.NetworkConfig{
		Driver: "bridge",
	}
	for _, service := range combinedCompose.Services {
		delete(service.Networks, "default")
		service.Networks[networkName] = &types.ServiceNetworkConfig{}
	}
	return combinedCompose, nil
}
