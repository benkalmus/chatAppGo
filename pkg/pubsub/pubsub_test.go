package pubsub_test

import (
    "os"
	"fmt"
	"sync"
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

// Helper Funcs


// Setups a pub sub, and a slice of rooms depending on cap
func testSetup(numRooms int) *pubsub.PubSub {
	ps := pubsub.NewPubSub()
	// var sliceOfRooms = make([]*pubsub.Room, numRooms)
	for i := 0; i < numRooms; i++ {
		ps.NewRoom(fmt.Sprintf("testRoom_%d", i))
	}
	return ps
}

func testTeardown(ps *pubsub.PubSub) {
	wg := sync.WaitGroup{}

	//close all rooms async
	ps.Rooms.Range(func(key, value interface{}) bool {
		wg.Add(1)
		room := value.(*pubsub.Room)
		go func() {
			ps.StopRoom(room)
			wg.Done()
		}()
		return true
	})
	wg.Wait()
	
}
