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
	composeFiles = append(composeFiles, Compose{Name: rootComposeFile.Name, Path: rootComposeFile.Name, Active: false})
	for _, composeFile := range childComposeFiles {
		composeFiles = append(composeFiles, Compose{Name: composeFile.Name, Path: composeFile.Name, Active: false})
	}
	return c.Render(http.StatusOK, "compose", map[string]interface{}{})
}