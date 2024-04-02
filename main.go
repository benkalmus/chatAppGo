package main

import (
    "os"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func main() {
	// Configure logging to write to both a file and standard output
    file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatal().Err(err).Msg("Failed to open log file")
    }
    defer file.Close()
	
    // Create a multiwriter to write logs to both file and standard output
    multi := zerolog.MultiLevelWriter(file, zerolog.ConsoleWriter{Out: os.Stdout})

    // Set up ZeroLog with multiwriter
    log.Logger = zerolog.New(multi).With().Timestamp().Logger()

    // Example log messages
    log.Info().Msg("Starting chat application")

}
