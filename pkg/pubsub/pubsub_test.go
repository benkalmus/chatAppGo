package pubsub_test

import (
	"chat_app/internal/logger"
	. "chat_app/pkg/pubsub"
	"fmt"
	"os"
	"sync"

	// "sync"
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

// func TestPubSub(t *testing.T) {
// 	log.Info().Msgf("Starting test %v", t.Name())
// 	testPubSub := NewPubSub()
// 	testRoom := testPubSub.NewRoom("testRoom", nil)	//with default opts
// 	testSub := NewSubscriber("testSub")
// 	err := testSub.Subscribe(testRoom)
// 	assert.ErrorIs(t, err, nil)

// 	// Publish a message on room1
// 	testRoom.Publish("hello")

// 	//  Wait for one message
// 	message := <- testSub.Recv()
// 	log.Debug().Msgf("Sub Received: %v\n", message.Payload)
// 	assert.Equal(t, "hello", message.Payload)
// 	testPubSub.StopRoom(testRoom)
// 	for testPubSub.IsRoomOpen(testRoom) {}
// }

// func TestSubscribeUnsubscribe(t *testing.T) {
// 	// ps := &PubSub{}
// 	// room := &Room{}
// 	// sub := &Subscriber{}
// 	ps, room, sub := setupMinimal()
// 	sub.Subscribe(room)
// 	room.Publish("hello")
// 	msgs  := []Message{}
// 	done := make(chan struct{})
// 	go subRecvLoop(sub, &msgs, done)

// 	err := sub.Unsubscribe(room)
// 	assert.Equal(t, nil, err)
// 	room.Publish("there")
// 	<-done 
// 	assert.Equal(t, 1, len(msgs))
// 	assert.Equal(t, "hello", msgs[0].Payload)
// 	teardownMinimal(ps, room, sub)
// }


// func TestMultipleRooms(t *testing.T) {
// 	log.Info().Msgf("Starting test %v", t.Name())
// 	testPubSub := NewPubSub()
// 	testRoomA := testPubSub.NewRoom("testRoomA", nil)
// 	testRoomB := testPubSub.NewRoom("testRoomB", nil)
// 	testSubA := NewSubscriber("testSub")
// 	testSubB := NewSubscriber("testSub")	//since this is another room, we can have the same name Sub
// 	testSubA.Subscribe(testRoomA)
// 	testSubB.Subscribe(testRoomB)

// 	// Publish a message on room1
// 	MsgA := NewMessage("hello")
// 	MsgB := NewMessage("hello")
// 	testRoomA.Publish(MsgA)
// 	testRoomB.Publish(MsgB)

// 	//  Wait for one message
// 	go subRecvOnce(t, testSubA, MsgA)
// 	go subRecvOnce(t, testSubB, MsgB)
// 	time.Sleep(10 * time.Millisecond)	//TODO
// 	testPubSub.StopRoom(testRoomA)
// 	testPubSub.StopRoom(testRoomB)
// }

// func TestSubMultipleMessages(t *testing.T) {
// 	ps, room, sub := setupMinimal()

// 	m1 := NewMessage("hello")
// 	m2 := NewMessage("there")
// 	msgSlice := []Message{m1, m2}
// 	for _, m := range msgSlice {
// 		room.Publish(m)
// 	}
// 	// wait for published messages
// 	messageBuffer := []Message{}
// 	done := make(chan struct{})
// 	subRecvLoop(sub, &messageBuffer, done)
// 	// room.Publish("hello")
// 	// room.Publish("there")
// 	<-done
// 	assert.Equal(t, msgSlice, messageBuffer)
// 	teardownMinimal(ps, room, sub)
// }


// func TestTrySubscribeToClosedRoom(t *testing.T) {
// 	testPubSub := NewPubSub()
// 	testRoom := testPubSub.NewRoom("testRoom", nil)	//with default opts
// 	realSub := NewSubscriber("RealSub")
// 	testSub := NewSubscriber("testSub")
// 	realSub.Subscribe(testRoom)

// 	testPubSub.StopRoom(testRoom)
// 	// wait for room to close, the sub's channel will be closed when this happens
// 	_, open := <-realSub.Recv()
// 	assert.Equal(t, open, false)
// 	time.Sleep(100 * time.Millisecond)
// 	err := testSub.Subscribe(testRoom)
// 	assert.Error(t, err, "room is closed")
// }

// func TestPublishOnRoomWithSubThatCrashed(t *testing.T) {
// 	testPubSub := NewPubSub()
// 	testRoom := testPubSub.NewRoom("testRoom", nil)	//with default opts
// 	testSub := NewSubscriber("testSub")
//  	testSub.Subscribe(testRoom)
// 	// Publish a message on room1
// 	testRoom.Publish("hello")
// 	//  Wait for one message
// 	message := <- testSub.Recv()
// 	log.Debug().Msgf("Sub Received: %v\n", message.Payload)
	
// 	testSub = &Subscriber{} 

// 	// close(testRoom.subscribers.Load("testSub").(chan Message)) //close channel to simulate broken Sub
// 	testRoom.Publish("test")
// 	testSub.Subscribe(testRoom) // should work because this subscriber should no longer be subscribed

// }

// Some test ideas:

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

func TestStopRoom(t *testing.T) {
	pubSub := NewPubSub()
	room := pubSub.NewRoom("A", nil)
	errR := pubSub.StopRoom(room)
	assert.Equal(t, nil, errR)
	// tries to stop room twice
	errR = pubSub.StopRoom(room)
	if assert.Error(t, errR) {
		assert.Equal(t, fmt.Errorf("room not found"), errR)
	}

	msg := NewMessage("Hello there")
	err := room.Publish(msg)		
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Errorf("context canceled"), err)
	}
	pubSub.Stop()
}
func TestStopPubSub(t *testing.T){
	pubSub := NewPubSub()
	room := pubSub.NewRoom("A", nil)
	sub, err := room.NewSubscriber()
	assert.Equal(t, nil, err)
	pubSub.Stop()
	t0 := time.Now()
	for room.IsAlive() { 
		if time.Since(t0) > 1 * time.Second {
			t.Log("Room not closed")
			t.Fail()
		}
	}
	t0 = time.Now()
	for sub.IsAlive() {
		if time.Since(t0) > 1 * time.Second {
			t.Log("Sub not stopped")
			t.Fail()
		}
	}
	pubSub.Stop()
}

func TestSubscriberRecv(t *testing.T) {
	pubSub := NewPubSub()
	// opts := RoomOpts{}
	room := pubSub.NewRoom("A", nil)
	sub, err := room.NewSubscriber()
	assert.Equal(t, nil, err)
	msg := NewMessage("Hello there")
	errp := room.Publish(msg)	//non blocking
	assert.Equal(t, nil, errp)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for m := range sub.Recv() {
			assert.Equal(t, msg, m)
		}
		t.Log("Sub closed")
		wg.Done()
	}()
	errs := room.Unsubscribe(sub)
	assert.Equal(t, nil, errs)
	pubSub.StopRoom(room)
	pubSub.Stop()
	wg.Wait()
}

// func TestMultipleSubscriberRecv(t *testing.T){
// 	pubSub := NewPubSub()
// 	room := pubSub.NewRoom("A", nil)
// 	sub, err := room.NewSubscriber()
// 	assert.Equal(t, nil, err)
// 	sub2,err2 := room.NewSubscriber()
// 	assert.Equal(t, nil, err2)
// 	go sub.RecvUntilClosed()
// 	go sub2.RecvUntilClosed()
// 	msgs := []Message{}
// 	for i := 0; i < 3; i++ {
// 		msg := NewMessage(fmt.Sprintf("msg_%d", i))
// 		err := room.Publish(msg)	//non blocking
// 		msgs = append(msgs, msg)
// 		assert.Equal(t, nil, err)
// 	}
// 	room.Stop()
// 	for sub.IsAlive() {}
// 	for sub2.isAlive() {}
// 	assert.Equal(t, msgs, sub.GetMessages())
// 	assert.Equal(t, msgs, sub2.GetMessages())
// 	pubSub.Stop()
// }





// Helper Funcs
// ===============================================


func setupMinimal() (ps *PubSub, room *Room, sub *Subscriber) {
	ps = NewPubSub()
	room = ps.NewRoom("room", nil)
	sub, err := room.NewSubscriber()
	if err != nil {
		log.Fatal().Msgf("Failed to create subscriber: %v", err)
	}
	return 
}

func teardownMinimal(ps *PubSub, room *Room, sub *Subscriber) {
	// room.Unsubscribe(sub)
	// ps.StopRoom(room)
	ps.Stop()
}

func subRecvOnce(t *testing.T, sub *Subscriber, expected Message) {
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

func subRecvLoop(sub *Subscriber, msgs *[]Message, done chan struct{}) {
	for {
		select {
		case message, open := <-sub.Recv(): 
			if !open {
				log.Debug().Msgf("Sub '%v' room closed '%v'", sub.Id, sub.Room)
				close(done)
				return
			}

			log.Debug().Msgf("Sub '%v' Received '%v'", sub.Id, message.Payload)
			m := message
			*msgs = append(*msgs, m)
		case <-time.After(1 * time.Second):
			close(done)
			return 
		}
	}
}



