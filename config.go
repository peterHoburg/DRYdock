package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

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
	viper.SetDefault("REMOVE_ORPHANS", false)
	viper.SetDefault("ALWAYS_RECREATE_DEPS", false)
	viper.SetDefault("STOP_ALL_CONTAINERS_BEFORE_RUNNING", "")

	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("PORT", "1994")
	viper.SetDefault("CUSTOM_COMPOSE_COMMAND", "")
	viper.SetDefault("COMPOSE_FILE_NAME_OVERRIDE", "")
	viper.SetDefault("PRE_RUN_COMMAND", "")
	viper.SetDefault("ENV_VAR_FORMAT", ".env-[[ENVIRONMENT]]")
	viper.SetDefault("ENV_VAR_FILE_SETUP_COMMAND", "")
	viper.SetDefault("", "")

	viper.SetConfigName("drydock")        // name of config file (without extension)
	viper.SetConfigType("yaml")           // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/drydock/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.drydock") // call multiple times to add many search paths
	viper.AddConfigPath(".")              // optionally look for config in the working directory
	err := viper.ReadInConfig()           // Find and read the config file
	if err != nil {                       // Handle errors reading the config file
		log.Error().Err(err).Msg("No config file loaded")
	} else {
		log.Info().Msg(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))
	}
	viper.SetEnvPrefix("DRYDOCK_")
	viper.AutomaticEnv()
}
