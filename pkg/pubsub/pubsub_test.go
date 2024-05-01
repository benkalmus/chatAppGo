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

// Some test ideas:

//X Subscriber tries to subscribe to a closed room or non existing
//X Subscriber tries to receive message on closed room
// Subscriber can only subscribe to one room
// Subscriber can unsub then sub to another room
//X Unsubscribed sub no longer receives messages
//X Subscriber no longer receiving messages on closed room 
//X If a subscriber crashes, dies or is no longer available, it should be removed from the room and publishing should still work to other subscribers

// Publishing actions should be non-blocking, 
// Publish on a closed or non existing room 
// Published messages arrive in order, verify with timestamp 

func Test_start_stop_room(t *testing.T) {
	pubSub := NewPubSub()
	room := pubSub.NewRoom("A", nil)
	errR := pubSub.StopRoom(room)
	assert.Nil(t, errR)
	// tries to stop room twice
	errR = pubSub.StopRoom(room)
	assert.Error(t, errR) 
	assert.Equal(t, fmt.Errorf("room not found"), errR)
	
	_, err := pubSub.FindRoom("A")
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("room not found"), err)
	pubSub.Stop()
}
func Test_start_stop_pubsub(t *testing.T){
	pubSub := NewPubSub()
	room := pubSub.NewRoom("A", nil)
	sub, err := room.NewSubscriber()
	assert.Nil(t, err)
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

func Test_subscriber_recv(t *testing.T) {
	pubSub := NewPubSub()
	// opts := RoomOpts{}
	room := pubSub.NewRoom("A", nil)
	sub, err := room.NewSubscriber()
	assert.Nil(t, err)
	msg := NewMessage("Hello there")
	errp := room.Publish(msg)
	assert.Nil(t, errp)
	wg := sync.WaitGroup{}
	wg.Add(1)
	msgChan := make(chan Message, 1)
	go func(ch chan Message) {
		m := <- sub.Recv()
		ch <- m
		t.Log("Sub closed")
		close(ch)
		wg.Done()
	}(msgChan)
	wg.Wait()
	errs := room.Unsubscribe(sub)
	assert.Nil(t, errs)
	pubSub.StopRoom(room)
	pubSub.Stop()
	select {
	case m := <-msgChan:
		assert.Equal(t, msg, m)
	case <-time.After(time.Second):
		t.Fail()
	}
}

func Test_multiple_subscriber_recv(t *testing.T){
	pubSub := NewPubSub()
	room := pubSub.NewRoom("A", nil)
	sub, err := room.NewSubscriber()
	assert.Nil(t, err)
	sub2, err2 := room.NewSubscriber()
	assert.Nil(t, err2)
	msgBuffer, msgBuffer2 := []Message{}, []Message{}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go RecvUntilClosed(t, sub, &msgBuffer, &wg)
	go RecvUntilClosed(t, sub2, &msgBuffer2, &wg)
	msgs := []Message{}
	for i := 0; i < 3; i++ {
		msg := NewMessage(fmt.Sprintf("msg_%d", i))
		err := room.Publish(msg)
		assert.Nil(t, err)
		msgs = append(msgs, msg)
	}
	room.Stop()
	// wait for all messages to arrive
	wg.Wait()
	assert.Equal(t, msgs, msgBuffer)
	assert.Equal(t, msgs, msgBuffer2)
	pubSub.Stop()
}

func Test_unsubscribe_no_longer_receives(t *testing.T) {
	pubSub := NewPubSub()
	room := pubSub.NewRoom("A", nil)
	sub, err := room.NewSubscriber()
	assert.Nil(t, err)
	errUns := room.Unsubscribe(sub)
	assert.Nil(t, errUns)

	wg := sync.WaitGroup{}
	wg.Add(1)
	msgBuffer := []Message{}
	go RecvUntilClosed(t, sub, &msgBuffer, &wg)

	msg := NewMessage("Hello there")
	room.Publish(msg)
	room.Publish(msg)
	
	pubSub.StopRoom(room)
	pubSub.Stop()
	wg.Wait()
	assert.Empty(t, msgBuffer)
}

func Test_non_consuming_subscriber_removed_from_room(t *testing.T) {
	pubSub := NewPubSub()
	opts := &RoomOpts{
		PublishChanBuffer: 	1,
		SubChanBuffer: 		1,		// channel will be blocking after 1 message
		SubscriberTimeoutMs: 0,		//very short kick time
	}
	room := pubSub.NewRoom("A", opts)
	inactiveSub, err := room.NewSubscriber()
	assert.Nil(t, err)
	activeSub, err := room.NewSubscriber()
	assert.Nil(t, err)

	msg := NewMessage("Hello there")
	msgs := []Message{msg, msg}
	room.Publish(msg, msg)
	
	wg := sync.WaitGroup{}
	wg.Add(1)
	msgBuffer := []Message{}
	go RecvUntilClosed(t, activeSub, &msgBuffer, &wg)

	time.Sleep(5 * time.Millisecond)
	// inactive sub should be removed from the room
	alive := inactiveSub.IsAlive()
	assert.Equal(t, false, alive)
	pubSub.Stop()
	wg.Wait()
	// active sub should have received all messages
	assert.Equal(t, msgs, msgBuffer)
}

func Test_subscribe_dead_room(t *testing.T) {
	pubSub := NewPubSub()
	opts := &RoomOpts{
		PublishChanBuffer: 	1,
		SubChanBuffer: 		1,		// channel will be blocking after 1 message
	}
	room := pubSub.NewRoom("A", opts)
	pubSub.StopRoom(room)
	sub, err := room.NewSubscriber()
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("room is closed"), err)
	assert.Nil(t, sub)
	pubSub.Stop()
}

//TODO message counter test

// Helper Funcs
// ===============================================

func RecvUntilClosed(t *testing.T, sub *Subscriber, buffer *[]Message, wg *sync.WaitGroup) {
	for msg := range sub.Recv() {
		log.Debug().Msgf("Sub '%v' Received '%v'", sub.Id, msg.Payload)
		*buffer = append(*buffer, msg)
	}
	wg.Done()
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



