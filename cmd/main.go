package main

import (
	//	"fmt"
	// "os"
	"time"	
	"github.com/rs/zerolog/log"
	"chat_app/internal/logger"
	// internal packages
	"chat_app/internal/message_broker"
)

// Main
// ========================================

func main() {
	logger.Init("info")
	log.Info().Msg("Starting Chat App.")

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
	err := broker.Publish("test", "This message won't be received.")
	if err != nil {
		log.Info().Msgf("Failed to publish message: %v\n", err)
	}
	time.Sleep(time.Second * 3)

}
