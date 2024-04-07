package logger

import (
    "os"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

)

// var Logger zerolog.Logger

func Init(logLevel string) {
    // Parse the log level
    level, err := zerolog.ParseLevel(logLevel)
    if err != nil {
		// default level
        level = zerolog.InfoLevel
    }
	zerolog.SetGlobalLevel(level)

	//TODO setup log rollover
	// "gopkg.in/natefinch/lumberjack.v2"
	// Configure logging to write to both a file and standard output
	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open log file")
	}
	defer file.Close()

	// Create a multiwriter to write logs to both file and standard output
	multi := zerolog.MultiLevelWriter(file, zerolog.ConsoleWriter{Out: os.Stdout})

	// Set up ZeroLog with multiwriter
	log.Logger = zerolog.New(multi).With().Timestamp().Caller().Logger()
}

func InitTesting() {
	
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	// Set up ZeroLog 
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Caller().Logger()

}