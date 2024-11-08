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

func CombineComposeFiles(composeFiles []*types.Project) (*types.Project, error) {
	project, composeFiles := composeFiles[0], composeFiles[1:]
	for _, c := range composeFiles {
		for k, v := range c.Services {
			project.Services[k] = v
		}
	}
	return project, nil
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
