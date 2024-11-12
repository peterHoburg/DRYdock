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
	form, err := c.FormParams()
	if err != nil {
		return err
	}

	var composeFiles []string
	// TODO add env type to each(?) row.
	for k, v := range form {
		if v[0] == "on" {
			composeFiles = append(composeFiles, k)
		}
	}
	return nil
}
