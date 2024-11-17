package composeApi

import (
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/rs/zerolog/log"

	"drydock/internal"
)

type RunReturnData struct {
	Output     string
	LogCommand string
	Error      error
}

func handleErr(c echo.Context, err error) error {
	c.Response().Header().Add("HX-Retarget", "#errors")
	c.Response().Header().Add("HX-Reswap", "innerHTML")
	return c.Render(http.StatusOK, "errors", err)
}

func Get(c echo.Context) error {
	rootComposeFile, childComposeFiles, err := internal.GetAllComposeFiles()
	var composeFiles []internal.Compose

	if err != nil {
		log.Error().Err(err).Msg("")
		return handleErr(c, err)
	}

	composeFiles = append(composeFiles, internal.Compose{Name: "Root", Path: rootComposeFile.Project.WorkingDir, Active: internal.Pointer(false)})
	for _, composeFile := range childComposeFiles {
		composeFiles = append(composeFiles, internal.Compose{Name: composeFile.Name, Path: composeFile.Project.WorkingDir, Active: internal.Pointer(false)})
	}
	return c.Render(http.StatusOK, "containerRows", composeFiles)
}

func Run(c echo.Context) error {
	var defaultEnvironmentSelect string
	var environment string
	composeRunData := internal.ComposeRunData{}
	form, err := c.FormParams()
	if err != nil {
		log.Error().Err(err).Msg("")
		return handleErr(c, err)
	}

	for k, v := range form {
		if k == "defaultEnvironmentSelect" {
			defaultEnvironmentSelect = v[0]
			continue
		}
		if k == "removeOrphans" {
			composeRunData.RemoveOrphans = true
			continue
		}
		if k == "alwaysRecreateDeps" {
			composeRunData.AlwaysRecreateDeps = true
			continue
		}
		if k == "stopAllContainersBeforeRunning" {
			composeRunData.StopAllContainersBeforeRunning = true
			continue
		}
		if k == "customComposeCommand" {
			composeRunData.CustomComposeCommand = v[0]
			continue
		}
		if k == "composeFileNameOverride" {
			composeRunData.ComposeFileNameOverride = v[0]
			continue
		}
	}

	for k, v := range form {
		if (len(v) > 1 && v[1] == "on") || k == "rootComposeFile" {
			if v[0] == "default" {
				environment = defaultEnvironmentSelect
			} else {
				environment = v[0]
			}
			if k == "rootComposeFile" {
				composeRunData.ComposeFiles = append(composeRunData.ComposeFiles, &internal.Compose{
					Path:        v[0] + "/docker-compose.yml",
					Active:      internal.Pointer(true),
					Environment: &environment,
				})
				continue
			}
			composeRunData.ComposeFiles = append(composeRunData.ComposeFiles, &internal.Compose{
				Path:        k + "/docker-compose.yml",
				Active:      internal.Pointer(true),
				Environment: &environment,
			})
		}
	}
	composeRunData.ProjectName = fmt.Sprintf("project-%d", time.Now().Unix())
	composeRunData.NetworkName = fmt.Sprintf("network-%d", time.Now().Unix())
	composeRunData.EnvFilePath = "/home/peter/GolandProjects/DRYdock/testdata/example-repo-setup/.example-env-vars" // TODO generate the file path based on env that is being run
	composeRunData.NewDockerComposePath, err = filepath.Abs(fmt.Sprintf("docker-compose-%d.yml", time.Now().Unix()))
	if err != nil {
		log.Error().Err(err).Msg("")
		return handleErr(c, err)
	}

	composeRunDataReturn, err := internal.ComposeFilesToRunCommand(composeRunData)
	if err != nil {
		log.Error().Err(err).Msg("")
		return handleErr(c, err)
	}

	combinedComposeFileYaml, err := composeRunDataReturn.ComposeFile.Project.MarshalYAML()
	if err != nil {
		log.Error().Err(err).Msg("")
		return handleErr(c, err)
	}

	err = internal.WriteComposeFile(composeRunData.NewDockerComposePath, combinedComposeFileYaml)
	if err != nil {
		log.Error().Err(err).Msg("")
		return handleErr(c, err)
	}

	cmd := exec.Command("docker", composeRunDataReturn.Command...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if output == nil {
			log.Error().Err(err).Msg(fmt.Sprintf("Error running docker compose command: %s", composeRunDataReturn.Command))
			return handleErr(c, err)
		}
		log.Error().Err(err).Msg(fmt.Sprintf("Error running docker compose command: %s \nOutput: %s", composeRunDataReturn.Command, string(output)))
		return c.Render(http.StatusOK, "run", RunReturnData{Error: err, Output: string(output), LogCommand: "ERROR"})
	}
	log.Info().Msg(string(output))
	return c.Render(http.StatusOK, "run", RunReturnData{Output: string(output), LogCommand: fmt.Sprintf("docker compose -f %s logs -t -f ", composeRunDataReturn.ComposeFile.Path)})
}
