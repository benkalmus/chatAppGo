package pubsub

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Constant defs
type ActionType int32

const (
	//defaults
	BUFFER_ACTIONS     = 10
	BUFFER_PUBLISH     = 50
	BUFFER_SUB_CHANNEL = 50

	idHashLength = 8
)

// ========================================
// Types
// ========================================
type PubSub struct {
	// set of room ptrs
	Rooms map[string]*Room
	// maybe
	// Rooms map[*Room]struct{}

	// TODO: access to rooms
	mutex sync.Mutex
}

// func (r1 *Room) Equals(r2 *Room) bool {
//     return r1.name == r2.name
// }
// func (r1 *Room) NotEquals(r2 *Room) bool {
//     return !r1.Equals(r2)
// }

type Room struct {
	Name string
	// for async publishing of messages
	PublishChan chan Message
	//for async buffering of sub/unsub actions
	ActionsChan chan Action
	stopChan    chan bool // for closing the room gracefully
	//TODO, store channels instead?
	subscribers sync.Map // map[Subscriber]struct{}
}

type Subscriber struct {
	id   string
	room *Room // pointer to Room this Subscriber belongs to
	//TODO, convert this to receive-only channel
	// this can be done by storing channels in the Room instead of Subscribers. Then when a sub subscribes to a room, a channel is created for it with and converted to receive only with (<-chan Message)(RecvChannel)
	RecvChan chan Message
}

type Message struct {
	// unique msg id
	id        []byte
	timestamp time.Time
	Payload   interface{}
}

// Actions
// controling subscribers
// ========================================
type Action interface {
	updateSubs(*Room)
}

type SubscribeAction struct {
	sub        Subscriber
	statusChan *chan bool
}
type UnsubscribeAction struct {
	sub        Subscriber
	statusChan *chan bool
}

// ========================================
// API
// ========================================

// PubSub
func NewPubSub() *PubSub {
	return &PubSub{
		Rooms: make(map[string]*Room),
	}
}

// Room
// ========================================
func NewRoom(name string) *Room {
	return &Room{
		Name: name,
		// TODO, add optional buffer values
		PublishChan: make(chan Message, BUFFER_PUBLISH),
		ActionsChan: make(chan Action, BUFFER_ACTIONS),
		stopChan:    make(chan bool, 1),
		subscribers: sync.Map{},
	}
}

func (r *Room) Subscribe(sub *Subscriber) chan bool {
	// create a channel that wil be used to inform the Subscriber when the action succeeded
	respChan := make(chan bool)
	r.ActionsChan <- &SubscribeAction{sub: *sub, statusChan: &respChan}
	return respChan
}

func (r *Room) Unsubscribe(sub *Subscriber) chan bool {
	respChan := make(chan bool)
	r.ActionsChan <- &UnsubscribeAction{sub: *sub, statusChan: &respChan}
	return respChan
}
func (r *Room) Publish(msgs ...interface{}) error {
	for _, msg := range msgs {
		switch msg.(type) {
		case Message:
			r.PublishChan <- msg.(Message)
		case string:
			r.PublishChan <- NewMessage(msg.(string))
		default:
			log.Error().Msgf("Attempt to publish unsupported msg %T", msg)
			return fmt.Errorf("Unsupported publish type %T", msg)
		}
	}
	return nil
}

func NewMessage(text string) Message {
	//create a unique random for id
	bytes := make([]byte, idHashLength)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return Message{bytes, time.Now(), text}
}

func (r *Room) StartRoom() {
	go r.Run()
}

func (r *Room) StopRoom() {
	log.Info().Msgf("Sending stop request to room %v", r.Name)
	r.stopChan <- true
}

// Subscriber
// ========================================

func NewSubscriber(id string) *Subscriber {
	return &Subscriber{
		id:       id,
		RecvChan: make(chan Message, BUFFER_SUB_CHANNEL),
		room:     nil,
	}
}

func (s *Subscriber) Recv() <-chan Message {
	return s.RecvChan
}

func (s *Subscriber) Stop() {
	StatusChan := s.room.Unsubscribe(s)
	// wait for StatusChannel to close
	<-StatusChan
	log.Info().Msgf("Unsubscribed %v from room %v", s.id, s.room.Name)
}

// Room loop: handles incoming publish and subscription requests
func (r *Room) Run() {
	defer r.cleanup()

	log.Info().Msgf("Starting room %v", r.Name)
	for {
		select {
		case msg := <-r.PublishChan:
			log.Debug().Msgf("Room %v received msg %v", r.Name, msg.Payload)
			handlePublish(r, msg)
		case act := <-r.ActionsChan:
			act.updateSubs(r)
		case <-r.stopChan:
			log.Info().Msgf("Received close request for room %v", r.Name)
			return
		}
	}
}

// Internal func
// ========================================

func (s *SubscribeAction) updateSubs(r *Room) {
	log.Info().Msgf("Subscribing new %v to room %v", s.sub.id, r.Name)
	r.subscribers.Store(s.sub, struct{}{})
	*s.statusChan <- true
	close(*s.statusChan)
}
func (s *UnsubscribeAction) updateSubs(r *Room) {
	log.Info().Msgf("Unsubscribing %v from room %v", s.sub.id, r.Name)
	r.subscribers.Delete(s.sub)
	*s.statusChan <- true
	close(s.sub.RecvChan)
	close(*s.statusChan)
}

func handlePublish(r *Room, msg Message) {
	log.Info().Msgf("Publishing to room %v", r.Name)
	//iterate over sync.Map in r.subscribers
	r.subscribers.Range(func(sub, _ interface{}) bool {
		castSub := sub.(Subscriber)
		//todo : async go
		castSub.RecvChan <- msg
		return true
	})

}

func (r *Room) cleanup() {
	// inform subscribers of room stop
	r.subscribers.Range(func(sub, _ interface{}) bool {
		castSub := sub.(Subscriber)
		close(castSub.RecvChan)
		return true
	})
	// Close all channels
	close(r.ActionsChan)
	close(r.PublishChan)
	close(r.stopChan)
	log.Info().Msgf("Cleaned and closed room %v", r.Name)
}
