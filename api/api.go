package api

import (
	"embed"
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

// viewsFs holds our static web server content.
//
//go:embed views/index.html
var viewsFs embed.FS

//go:embed static/*
var staticFs embed.FS

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseFS(viewsFs, "views/*.html")),
	}
}

type Count struct {
	Count int
}

func Start() {
	log.Info().Msg("Starting Drydock API")

	e := echo.New()

	// Middleware
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Output: log.Logger,
	}))
	e.Use(middleware.Recover())
	count := Count{Count: 0}
	e.Renderer = newTemplate()

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", count)
	})

	e.StaticFS("/static", staticFs)
	e.FileFS("/favicon.ico", "favicon.ico", staticFs)
	e.GET("/compose", ComposeGet)
	e.POST("/compose/run", ComposeRun)
	e.Logger.Fatal(e.Start(":1323"))
}
