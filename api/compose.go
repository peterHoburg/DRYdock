package api

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
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

func ComposeGet(c echo.Context) error {
	rootComposeFile, childComposeFiles, err := internal.GetAllComposeFiles()
	var composeFiles []internal.Compose

	if err != nil {
		log.Error().Err(err).Msg("")
		return handleErr(c, err)
	}

	composeFiles = append(composeFiles, internal.Compose{Name: "Root", Path: rootComposeFile.Project.WorkingDir, Active: false})
	for _, composeFile := range childComposeFiles {
		composeFiles = append(composeFiles, internal.Compose{Name: composeFile.Name, Path: composeFile.Project.WorkingDir, Active: false})
	}
	return c.Render(http.StatusOK, "containerRows", composeFiles)
}

func ComposeRun(c echo.Context) error {
	composeRunData := internal.ComposeRunData{}
	form, err := c.FormParams()
	if err != nil {
		log.Error().Err(err).Msg("")
		return handleErr(c, err)
	}

	composeRunData, err = composeRunData.LoadFromForm(form)
	if err != nil {
		log.Error().Err(err).Msg("")
		return handleErr(c, err)
	}
	composeRunData = composeRunData.ReplacePlaceholders()

	if composeRunData.PreRunCommand != "" {
		command := strings.Split(composeRunData.PreRunCommand, " ")

		cmd := exec.Command(command[0], command[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("Error running pre-run command: %s \nOutput: %s", composeRunData.PreRunCommand, string(output)))
			return handleErr(c, err)
		}
		log.Info().Msg(string(output))
	}

	composeRunData.ProjectName = fmt.Sprintf("project-%d", time.Now().Unix())
	composeRunData.NetworkName = fmt.Sprintf("network-%d", time.Now().Unix())

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
