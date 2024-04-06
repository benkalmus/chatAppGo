package main

import (
	//	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
	// internal packages
	"chat_app/internal/message_broker"
)

// Main
// ========================================

func main() {
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
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()

	// Example log messages
	log.Info().Msg("Starting chat application")

	broker := message_broker.NewBroker()

	subscriber := broker.Subscribe("test")
	go func() {
		for {
			select {
			case msg, ok := <-subscriber.Channel:
				if !ok {
					log.Info().Msg("Subscriber channel closed.")
					return
				}
				log.Info().Msgf("Received: %v\n", msg)
			case <-subscriber.Unsubscribe:
				log.Info().Msg("Unsubscribed.")
				return
			}
		}
	}()

	broker.Publish("test", "Hello, World!")
	broker.Publish("test", "This is a test message.")

	time.Sleep(2 * time.Second)
	broker.Unsubscribe("test", subscriber)

	broker.Publish("test", "This message won't be received.")
	err = broker.Publish("test", "This message won't be received.")
	if err != nil {
		log.Info().Msgf("Failed to publish message: %v\n", err)
	}
	time.Sleep(time.Second * 3)

}
