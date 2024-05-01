package main

import (
	//"fmt"
	// "os"
	// "time"	
	"github.com/rs/zerolog/log"
	"chat_app/internal/logger"
	// internal packages
	// "chat_app/pkg/pubsub"
)

// Main
// ========================================


func main() {
	logger.Init("info")
	log.Info().Msg("Starting Chat App.")

	// TODO, currently nothing is happening here. 
	// See pubsub_test.go

}
