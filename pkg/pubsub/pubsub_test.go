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
	testPubSub := pubsub.NewPubSub()
	testRoom := testPubSub.NewRoom("testRoom")
	testRoom.StartRoom()
	testSub := pubsub.NewSubscriber("testSub")
	statusChan := testRoom.Subscribe(testSub)

	if status := <-statusChan; !status {
		t.Errorf("subscription failed")
	}
	// Publish a message on room1
	testRoom.Publish("hello")

	//  Wait for one message
	message := <- testSub.Recv()
	log.Debug().Msgf("Sub Received: %v\n", message.Payload)
	assert.Equal(t, "hello", message.Payload)
}
