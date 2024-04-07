package pubsub_test

import (
	"chat_app/internal/logger"
	"chat_app/pkg/pubsub"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
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
	testRoom := testPubSub.NewRoom("testRoom", nil)	//with default opts
	testSub := pubsub.NewSubscriber("testSub")
	statusChan := testRoom.Subscribe(testSub)

	if testSub.AwaitSubscribe(statusChan, testRoom) != nil {
		t.Errorf("subscription failed")
	}
	// Publish a message on room1
	testRoom.Publish("hello")

	//  Wait for one message
	message := <- testSub.Recv()
	log.Debug().Msgf("Sub Received: %v\n", message.Payload)
	assert.Equal(t, "hello", message.Payload)
}

func TestMultipleRooms(t *testing.T) {
	log.Info().Msgf("Starting test %v", t.Name())
	testPubSub := pubsub.NewPubSub()
	testRoomA := testPubSub.NewRoom("testRoomA", nil)
	testRoomB := testPubSub.NewRoom("testRoomB", nil)
	testSubA := pubsub.NewSubscriber("testSub")
	statusChanA := testRoomA.Subscribe(testSubA)
	testSubB := pubsub.NewSubscriber("testSub")	//since this is another room, we can have the same name Sub
	statusChanB := testRoomB.Subscribe(testSubB)

	if testSubA.AwaitSubscribe(statusChanA, testRoomA) != nil {
		t.Errorf("subscription A failed")
	}
	if testSubB.AwaitSubscribe(statusChanB, testRoomB) != nil {
		t.Errorf("subscription B failed")
	}
	// Publish a message on room1
	MsgA := pubsub.NewMessage("hello")
	MsgB := pubsub.NewMessage("hello")
	testRoomA.Publish(MsgA)
	testRoomB.Publish(MsgB)

	//  Wait for one message
	go subRecvOnce(t, testSubA, MsgA)
	go subRecvOnce(t, testSubB, MsgB)
}

func TestSubMultipleMessages(t *testing.T) {
	ps := testSetup(1)
	roomAny, ok := ps.Rooms.Load("testRoom_0")
	if !ok {
		t.Errorf("room 'testRoom_0' not found")
	}
	room, ok := roomAny.(*pubsub.Room)
	if !ok {
		t.Errorf("type assertion on room failed")
	}
	sub := pubsub.NewSubscriber("testSub")
	statusChan := room.Subscribe(sub)
	if sub.AwaitSubscribe(statusChan, room) != nil {
		t.Errorf("subscription failed")
	}

	m1 := pubsub.NewMessage("hello")
	m2 := pubsub.NewMessage("there")
	msgSlice := []pubsub.Message{m1, m2}
	for _, m := range msgSlice {
		room.Publish(m)
	}
	//TODO, remove sleep by waiting for all messages to arrive before we stop the pubsub
	time.Sleep(10 * time.Millisecond)
	ps.StopRoom(room)
	messageBuffer := []pubsub.Message{}
	subRecvLoop(sub, &messageBuffer)
	// room.Publish("hello")
	// room.Publish("there")
	assert.Equal(t, msgSlice, messageBuffer)
}

// Some test criteria:
// Subscriber tries to subscribe to a closed room or non existing
// Subscriber tries to receive message on closed room
// Subscriber can only subscribe to one room
// Subscriber can unsub then sub to another room
// Unsubscribed sub no longer receives messages
// Subscriber no longer receiving messages on closed room 
// If a subscriber crashes, dies or is no longer available, it should be removed from the room and publishing should still work to other subscribers
// Publishing actions should be non-blocking, 

// Publish on a closed or non existing room 
// Published messages arrive in order, verify with timestamp 

// Helper Funcs
// ===============================================

// Setups a pub sub, and a slice of rooms depending on cap
func testSetup(numRooms int) *pubsub.PubSub {
	ps := pubsub.NewPubSub()
	// var sliceOfRooms = make([]*pubsub.Room, numRooms)
	for i := 0; i < numRooms; i++ {
		ps.NewRoom(fmt.Sprintf("testRoom_%d", i), nil)
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

func subRecvOnce(t *testing.T, sub *pubsub.Subscriber, expected pubsub.Message) {
	select  {
	case msg := <- sub.Recv():
		log.Debug().Msgf("Sub '%v' Received '%v'", sub.Id, msg.Payload)
		assert.Equal(t, expected.Payload, msg.Payload)
		assert.Equal(t, expected.Id, msg.Id)
		return 
	case <-time.After(1 * time.Second):
		t.Errorf("subscriber did not receive message")
	}
}

func subRecvLoop(sub *pubsub.Subscriber, msgs *[]pubsub.Message) {
	for message := range sub.Recv() {
		log.Debug().Msgf("Sub '%v' Received '%v'", sub.Id, message.Payload)
		m := message
		*msgs = append(*msgs, m)
	}
	log.Debug().Msgf("Sub '%v' room closed '%v'", sub.Id, sub.Room)
}
