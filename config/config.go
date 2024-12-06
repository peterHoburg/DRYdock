package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func getRootDir() string {
	//ex, err := os.Executable()
	//if err != nil {
	//	panic(err)
	//}
	// "." will give a better result for where the binary is run from that os.Executable()
	exPath, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	log.Info().Msg(fmt.Sprintf("ROOT_DIR: %s", exPath))
	return exPath
}

func InitLogger() {
	// TODO set logging level via config
	dirname, err := os.UserHomeDir()
	logDir := fmt.Sprintf("%s/.drydock/logs", dirname)
	err = os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		log.Fatal().Err(err).Msg("Error failed to make logDir")
	}
	logFile, err := os.OpenFile(fmt.Sprintf("%s/%d.log", logDir, time.Now().Unix()), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening file")
	}

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}
	multi := zerolog.MultiLevelWriter(consoleWriter, logFile)
	log.Logger = zerolog.New(multi).Level(zerolog.TraceLevel).With().Timestamp().Caller().Logger()
}

func LoadConfig() {
	rootDir := getRootDir()
	composeFileName, err := filepath.Abs(filepath.Join(rootDir, "docker-compose-[[TIMESTAMP]].yml"))
	if err != nil {
		log.Error().Err(err).Msg("Error getting composeFileName")
		composeFileName = "docker-compose-[[TIMESTAMP]].yml"
	}

	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("PORT", "1994")

	viper.SetDefault("REMOVE_ORPHANS", false)
	viper.SetDefault("ALWAYS_RECREATE_DEPS", false)
	viper.SetDefault("STOP_ALL_CONTAINERS_BEFORE_RUNNING", false)

	viper.SetDefault("ROOT_DIR", rootDir)
	viper.SetDefault("COMPOSE_FILE_REGEX", "^docker-compose\\.ya?ml$")
	viper.SetDefault("COMPOSE_COMMAND", "compose -f [[COMPOSE_FILE_NAME]] up --build -d")
	viper.SetDefault("COMPOSE_FILE_NAME", composeFileName)
	viper.SetDefault("PRE_RUN_COMMAND", "")
	viper.SetDefault("ENV_VAR_FORMAT", ".env-[[ENVIRONMENT]]")
	viper.SetDefault("ENV_VAR_FILE_SETUP_COMMAND", "")
	viper.SetDefault("VARIABLE_INTERPOLATION_OPTIONS", "")

	viper.SetConfigName("drydock")        // name of config file (without extension)
	viper.SetConfigType("yaml")           // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/drydock/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.drydock") // call multiple times to add many search paths
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("DRYDOCK_")
	// optionally look for config in the working directory
	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		log.Error().Err(err).Msg("No config file loaded")
	} else {
		log.Info().Msg(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))
	}
	viper.AutomaticEnv()
}
