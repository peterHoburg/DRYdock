package composeApi

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"

	"drydock/internal"
)

type RunReturnData struct {
	Output     string
	LogCommand string
	Error      error
}

func handleErr(c echo.Context, err error) error {
	log.Println(err)
	c.Response().Header().Add("HX-Retarget", "#errors")
	c.Response().Header().Add("HX-Reswap", "innerHTML")
	return c.Render(http.StatusOK, "errors", err)
}

func Get(c echo.Context) error {
	// TODO remove root from UI, But we need to find it in the run function
	rootComposeFile, childComposeFiles, err := internal.GetAllComposeFiles()
	if err != nil {
		return handleErr(c, err)
	}
	var composeFiles []internal.Compose
	composeFiles = append(composeFiles, internal.Compose{Name: "Root", Path: rootComposeFile.Project.WorkingDir, Active: internal.Pointer(false)})
	for _, composeFile := range childComposeFiles {
		composeFiles = append(composeFiles, internal.Compose{Name: composeFile.Name, Path: composeFile.Project.WorkingDir, Active: internal.Pointer(false)})
	}
	return c.Render(http.StatusOK, "containerRows", composeFiles)
}

func Run(c echo.Context) error {
	var defaultEnvironmentSelect string
	var composeFiles []*internal.Compose

	form, err := c.FormParams()
	if err != nil {
		return handleErr(c, err)
	}
	var environment string

	for k, v := range form {
		if k == "defaultEnvironmentSelect" {
			defaultEnvironmentSelect = v[0]
			continue
		}
	}
	// TODO handle when nothing is active
	for k, v := range form {
		if (len(v) > 1 && v[1] == "on") || k == "rootComposeFile" {
			if v[0] == "default" {
				environment = defaultEnvironmentSelect
			} else {
				environment = v[0]
			}
			if k == "rootComposeFile" {
				composeFiles = append(composeFiles, &internal.Compose{
					Path:        v[0] + "/docker-compose.yml",
					Active:      internal.Pointer(true),
					Environment: &environment,
				})
				continue
			}
			composeFiles = append(composeFiles, &internal.Compose{
				Path:        k + "/docker-compose.yml",
				Active:      internal.Pointer(true),
				Environment: &environment,
			})
		}
	}
	projectName := fmt.Sprintf("project-%d", time.Now().Unix())
	networkName := fmt.Sprintf("network-%d", time.Now().Unix())
	envFilePath := "/home/peter/GolandProjects/DRYdock/testdata/example-repo-setup/.example-env-vars" // TODO generate the file path based on env that is being run
	newDockerComposePath, err := filepath.Abs(fmt.Sprintf("docker-compose-%d.yml", time.Now().Unix()))
	if err != nil {
		log.Println(err)
	}

	composeRunData := internal.ComposeRunData{
		ComposeFiles:         composeFiles,
		ProjectName:          projectName,
		NetworkName:          networkName,
		EnvFilePath:          envFilePath,
		NewDockerComposePath: newDockerComposePath,
	}
	composeRunDataReturn, err := internal.ComposeFilesToRunCommand(composeRunData)
	if err != nil {
		return handleErr(c, err)
	}

	combinedComposeFileYaml, err := composeRunDataReturn.ComposeFile.Project.MarshalYAML()
	if err != nil {
		return handleErr(c, err)
	}

	err = internal.WriteComposeFile(composeRunData.NewDockerComposePath, combinedComposeFileYaml)
	if err != nil {
		return handleErr(c, err)
	}

	cmd := exec.Command("docker", composeRunDataReturn.Command...)
	output, err := cmd.CombinedOutput()
	println(string(output))

	if err != nil {
		if output == nil {
			return handleErr(c, err)
		}
		return c.Render(http.StatusOK, "run", RunReturnData{Error: err, Output: string(output), LogCommand: "ERROR"})
	}
	return c.Render(http.StatusOK, "run", RunReturnData{Output: string(output), LogCommand: fmt.Sprintf("docker compose -f %s logs -t -f ", composeRunDataReturn.ComposeFile.Path)})
}
