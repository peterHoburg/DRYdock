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
	composeFiles = append(composeFiles, Compose{Name: "Root", Path: rootComposeFile.Project.WorkingDir, Active: false})
	for _, composeFile := range childComposeFiles {
		composeFiles = append(composeFiles, Compose{Name: composeFile.Name, Path: composeFile.Project.WorkingDir, Active: false})
	}
	return c.Render(http.StatusOK, "containerRows", composeFiles)
}

func Run(c echo.Context) error {
	var defaultEnvironmentSelect string
	var composeRun []internal.Compose

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

	for k, v := range form {
		if len(v) > 1 && v[1] == "on" {
			if v[0] == "default" {
				environment = defaultEnvironmentSelect
			} else {
				environment = v[0]
			}
			composeRun = append(composeRun, internal.Compose{
				Path:        k + "/docker-compose.yml",
				Active:      internal.Pointer(true),
				Environment: &environment,
			})
		}
	}
	//rootComposeFile, childComposeFiles, err := internal.LoadAndOrganizeComposeFiles(childComposeFilePaths, projectName)
	//
	//var composeFiles []*types.Project
	//for _, compose := range composeRun {
	//	composeFile, err := internal.LoadComposeFile(compose.Path, "project")
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	composeFiles = append(composeFiles, composeFile)
	//}

	return nil
}
