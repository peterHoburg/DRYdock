package api

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
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

type IndexData struct {
	RootDir                      string
	ComposeFileRegex             string
	ComposeCommand               string
	ComposeFileName              string
	PreRunCommand                string
	EnvVarFileSetupCommand       string
	EnvVarFileFormat             string
	VariableInterpolationOptions string
}

func (i *IndexData) LoadFromViper() *IndexData {
	i.RootDir = viper.Get("ROOT_DIR").(string)
	i.ComposeCommand = viper.Get("COMPOSE_COMMAND").(string)
	i.ComposeFileName = viper.Get("COMPOSE_FILE_NAME").(string)
	i.ComposeFileRegex = viper.Get("COMPOSE_FILE_REGEX").(string)
	i.PreRunCommand = viper.Get("PRE_RUN_COMMAND").(string)
	i.EnvVarFileFormat = viper.Get("ENV_VAR_FILE_FORMAT").(string)
	i.EnvVarFileSetupCommand = viper.Get("ENV_VAR_FILE_SETUP_COMMAND").(string)
	i.VariableInterpolationOptions = viper.Get("VARIABLE_INTERPOLATION_OPTIONS").(string)
	return i
}

func Start() {
	log.Info().Msg("Starting Drydock API")
	indexData := IndexData{}
	indexData = *indexData.LoadFromViper()

	e := echo.New()

	// Middleware
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Output: log.Logger,
	}))
	e.Use(middleware.Recover())
	e.Renderer = newTemplate()

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", indexData)
	})

	e.StaticFS("/static", staticFs)
	e.FileFS("/favicon.ico", "static/favicon.png", staticFs)
	e.POST("/compose/list", ComposeGet)
	e.POST("/compose/run", ComposeRun)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", viper.Get("PORT").(string))))
}
