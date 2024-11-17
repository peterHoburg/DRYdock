package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger() {
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
	log.Logger = zerolog.New(multi).With().Timestamp().Caller().Logger()
}
