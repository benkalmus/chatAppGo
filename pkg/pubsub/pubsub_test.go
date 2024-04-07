package pubsub_test

import (
    "os"
	"chat_app/internal/logger"
    "github.com/rs/zerolog/log"
	"chat_app/pkg/pubsub"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMain(m *testing.M) {
    // Set up Zerolog for testing
    logger.InitTesting()
    // Run the tests
    os.Exit(m.Run())
}

func TestPubSub(t *testing.T) {
	log.Info().Msgf("Starting test %v", t.Name())
	//p := pubsub.NewPubSub()
	r := pubsub.NewRoom("room1")
	r.StartRoom()
	sub := pubsub.NewSubscriber("sub1")
	statusChan := r.Subscribe(sub)

	if status := <-statusChan; !status {
		t.Errorf("subscription failed")
	}
	// Publish a message on room1
	r.Publish("hello")

	//  Wait for one message
	message := <- sub.Recv()
	log.Debug().Msgf("Sub Received: %v\n", message.Payload)
	assert.Equal(t, "hello", message.Payload)
}
