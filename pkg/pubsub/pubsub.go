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
	Rooms sync.Map
}

type Room struct {
	Name string
	// for async publishing of messages
	PublishChan chan Message
	//for async buffering of sub/unsub actions
	ActionsChan chan Action
	StopChan    chan bool // for closing the room gracefully
	//TODO, store channels instead?
	subscribers sync.Map // map[Subscriber]struct{}
}

type RoomOpts struct {
	PublishChanBuffer int
	ActionsChanBuffer int 
}

type Subscriber struct {
	Id   string
	//TODO
	Room *Room // pointer to Room this Subscriber belongs to
	//TODO, convert this to receive-only channel
	// this can be done by storing channels in the Room instead of Subscribers. Then when a sub subscribes to a room, a channel is created for it with and converted to receive only with (<-chan Message)(RecvChannel)
	RecvChan chan Message
}

type Message struct {
	// unique msg id
	Id        []byte
	Timestamp time.Time
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
		Rooms: sync.Map{},
	}
}

// PubSub
// ========================================
func (ps *PubSub) NewRoom(name string, opts *RoomOpts) *Room {
	if opts == nil {
		opts = &RoomOpts{
			PublishChanBuffer: BUFFER_PUBLISH,
			ActionsChanBuffer: BUFFER_ACTIONS,
		}
	}

	room := &Room{
		Name: name,
		PublishChan: make(chan Message, opts.PublishChanBuffer),
		ActionsChan: make(chan Action, opts.ActionsChanBuffer),
		StopChan:    make(chan bool, 1),
		subscribers: sync.Map{},
	}
	ps.Rooms.Store(name,room)
	ps.startRoom(room)
	return room
}

// async
func (ps *PubSub) startRoom(r *Room) {
	go r.Run()
}

// sync
func (ps *PubSub) StopRoom(r *Room) error {
	_, ok := ps.Rooms.Load(r.Name)
	if !ok {
		return fmt.Errorf("room '%v' not found in pubsub '%v", r.Name, &ps)
	}
	log.Info().Msgf("Sending stop request to room '%v'", r.Name)
	r.StopChan <- true
	ps.Rooms.Delete(r.Name)
	return nil
}

// Room
// ========================================

// Room Goroutine loop: handles incoming publish and subscription requests
func (r *Room) Run() {
	defer r.cleanup()

	log.Info().Msgf("Starting room '%v'", r.Name)
	for {
		select {
		case msg := <-r.PublishChan:
			log.Debug().Msgf("publishing msg '%v' to '%v'", msg.Payload, r.Name,)
			handlePublish(r, msg)
		case act := <-r.ActionsChan:
			act.updateSubs(r)
		case <-r.StopChan:
			log.Info().Msgf("'%v' received close request", r.Name)
			return
		}
	}
}

func (r *Room) Subscribe(sub *Subscriber) chan bool {
	// create a channel that wil be used to inform the Subscriber when the action succeeded
	respChan := make(chan bool, 1)
	r.ActionsChan <- &SubscribeAction{sub: *sub, statusChan: &respChan}
	return respChan
}

func (r *Room) Unsubscribe(sub *Subscriber) chan bool {
	respChan := make(chan bool, 1)
	r.ActionsChan <- &UnsubscribeAction{sub: *sub, statusChan: &respChan}
	return respChan
}
func (r *Room) Publish(msgs ...interface{}) error {
	for _, msg := range msgs {
		switch msg := msg.(type) {
		case Message:
			r.PublishChan <- msg
		case string:
			r.PublishChan <- NewMessage(msg)
		default:
			log.Error().Msgf("Attempt to publish unsupported msg '%T'", msg)
			return fmt.Errorf("unsupported publish type '%T'", msg)
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

// Subscriber
// ========================================

func NewSubscriber(id string) *Subscriber {
	return &Subscriber{
		Id:       id,
		RecvChan: make(chan Message, BUFFER_SUB_CHANNEL),
		Room:     nil,
	}
}

func (s *Subscriber) Recv() <-chan Message {
	return s.RecvChan
}
func (s *Subscriber) AwaitSubscribe(respChan chan bool, r *Room) error {
	if _, ok := <-respChan; !ok {
		return fmt.Errorf("failed to subscribe '%v' to room '%v'", s.Id, r.Name)
	}
	s.Room = r
	log.Info().Msgf("Subscribed '%v' to room '%v'", s.Id, r.Name)
	return nil 
}


func (s *Subscriber) Stop() {
	StatusChan := s.Room.Unsubscribe(s)
	// wait for StatusChannel to close
	<-StatusChan
	log.Info().Msgf("Unsubscribed '%v' from room '%v'", s.Id, s.Room.Name)
}

// Internal func
// ========================================

func (s *SubscribeAction) updateSubs(r *Room) {
	log.Info().Msgf("Subscribing new '%v' to room '%v'", s.sub.Id, r.Name)
	r.subscribers.Store(s.sub, struct{}{})
	*s.statusChan <- true
	close(*s.statusChan)
}
func (s *UnsubscribeAction) updateSubs(r *Room) {
	log.Info().Msgf("Unsubscribing '%v' from room '%v'", s.sub.Id, r.Name)
	r.subscribers.Delete(s.sub)
	*s.statusChan <- true
	close(s.sub.RecvChan)
	close(*s.statusChan)
}

func handlePublish(r *Room, msg Message) {
	log.Info().Msgf("Publishing to room '%v'", r.Name)
	//iterate over sync.Map in r.subscribers
	wg := sync.WaitGroup{}
	r.subscribers.Range(func(sub, _ interface{}) bool {
		castSub := sub.(Subscriber)
		//todo : async go
		wg.Add(1)
		go func(){
			castSub.RecvChan <- msg
			wg.Done()
		}()
		return true
	})
	wg.Wait()
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
	close(r.StopChan)
	log.Info().Msgf("Cleaned and closed room '%v'", r.Name)
}
