package composeApi

import (
	"fmt"
	"log"
	"net/http"

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
		if len(v) > 1 && v[1] == "on" {
			if v[0] == "default" {
				environment = defaultEnvironmentSelect
			} else {
				environment = v[0]
			}
			composeFiles = append(composeFiles, &internal.Compose{
				Path:        k + "/docker-compose.yml",
				Active:      internal.Pointer(true),
				Environment: &environment,
			})
		}
	}
	combinedComposeFile, output, err := internal.RunComposeFiles(composeFiles)
	if err != nil {
		if output == nil {
			return handleErr(c, err)
		}
		return c.Render(http.StatusOK, "run", RunReturnData{Error: err, Output: string(*output), LogCommand: "ERROR"})
	}
	return c.Render(http.StatusOK, "run", RunReturnData{Output: string(*output), LogCommand: fmt.Sprintf("docker compose -f %s logs -t -f ", combinedComposeFile.Path)})
}
