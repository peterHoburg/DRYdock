package composeApi

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"drydock/internal"
)

type Compose struct {
	Name   string
	Path   string
	Active bool
}

type ComposeRun struct {
	Path        string
	Active      bool
	Environment string
}

func Get(c echo.Context) error {
	rootComposeFile, childComposeFiles, err := internal.GetAllComposeFiles()
	if err != nil {
		log.Fatal(err)
	}
	var composeFiles []Compose
	composeFiles = append(composeFiles, Compose{Name: "Root", Path: rootComposeFile.Project.WorkingDir, Active: false})
	for _, composeFile := range childComposeFiles {
		composeFiles = append(composeFiles, Compose{Name: composeFile.Name, Path: composeFile.Project.WorkingDir, Active: false})
	}
	return c.Render(http.StatusOK, "containerRows", composeFiles)
}

func Run(c echo.Context) error {
	var defaultEnvironmentSelect string
	var composeFiles []*internal.Compose

	form, err := c.FormParams()
	if err != nil {
		return err
	}
	var environment string

	for k, v := range form {
		if k == "defaultEnvironmentSelect" {
			defaultEnvironmentSelect = v[0]
			continue
		}
	}
	// TODO handle when nothing is active
	// TODO Should the root file be passed to the UI?? I don't think so
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
	err = internal.RunComposeFiles(composeFiles)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
