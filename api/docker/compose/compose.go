package composeApi

import (
	"fmt"
	"log"
	"net/http"
	"time"

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
	projectName := fmt.Sprintf("project-%d", time.Now().Unix())
	rootComposeFile, childComposeFiles, err := internal.GetAllComposeFiles(projectName)
	if err != nil {
		log.Fatal(err)
	}
	var composeFiles []Compose
	composeFiles = append(composeFiles, Compose{Name: "Root", Path: rootComposeFile.WorkingDir, Active: false})
	for _, composeFile := range childComposeFiles {
		composeFiles = append(composeFiles, Compose{Name: composeFile.Name, Path: composeFile.WorkingDir, Active: false})
	}
	return c.Render(http.StatusOK, "containerRows", composeFiles)
}

func Run(c echo.Context) error {
	var defaultEnvironmentSelect string
	var composeFiles []ComposeRun

	form, err := c.FormParams()
	if err != nil {
		return err
	}
	var environment string

	// TODO add env type to each(?) row.
	for k, v := range form {
		if k == "defaultEnvironmentSelect" {
			defaultEnvironmentSelect = v[0]
			continue
		}
	}

	for k, v := range form {
		if len(v) > 1 && v[1] == "on" {
			if v[0] == "default" {
				environment = defaultEnvironmentSelect
			} else {
				environment = v[0]
			}
			composeFiles = append(composeFiles, ComposeRun{
				Path:        k,
				Active:      true,
				Environment: environment,
			})
		}
	}
	print(defaultEnvironmentSelect)
	return nil
}
