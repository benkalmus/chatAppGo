package pubsub

import (
	"context"
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
	buffer_publish_chan = 50
	buffer_sub_channel  = 50

	idHashLength 		= 8
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
	ctx 		context.Context
	cancel 		context.CancelFunc
	subscribers sync.Map // map[*Subscriber](chan Message)
	MessageCount uint32
}

type RoomOpts struct {
	PublishChanBuffer int
	ActionsChanBuffer int 
}


type Subscriber struct {
	Id   		string
	// pointer to Room this Subscriber belongs to
	Room 		*Room 
	//Storing channels in the Room instead of Subscribers. Then when a sub subscribes to a room, a channel is created for it with and converted to receive only with (<-chan Message)(RecvChannel)
	RecvChan 	<-chan Message
	isAlive		bool
	mu 			sync.Mutex
}

type Message struct {
	// unique msg id
	Id        []byte
	Timestamp time.Time
	Payload   interface{}
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

func (ps *PubSub) Stop() {
	ps.Rooms.Range(func(key, value interface{}) bool {
		ps.StopRoom(value.(*Room))
		return true
	})
}

// PubSub
// ========================================
func (ps *PubSub) NewRoom(name string, opts *RoomOpts) *Room {
	if opts == nil {
		opts = &RoomOpts{
			PublishChanBuffer: buffer_publish_chan,
		}
	}

    ctx, cancel := context.WithCancel(context.Background())
	room := &Room{
		Name: name,
		PublishChan: 	make(chan Message, opts.PublishChanBuffer),
		ctx:    		ctx,
		cancel:			cancel,
		subscribers: 	sync.Map{},
	}
	ps.Rooms.Store(name, room)
	ready := make(chan struct{})
	go room.Run(ready)
	// wait for room to enter loop
	<-ready
	return room
}

func (ps *PubSub) StopRoom(r *Room) error {
	_, ok := ps.Rooms.Load(r.Name)
	if !ok {
		return fmt.Errorf("room not found")
	}
	log.Info().Msgf("Sending stop request to room '%v'", r.Name)
	r.cancel()
	//for r.IsAlive() {}
	log.Debug().Msgf("Waiting for room '%v' to close", r.Name)
	ps.Rooms.Delete(r.Name)
	return nil
}

func (r *Room) IsAlive() bool {
	select {
	case <-r.ctx.Done():
		return false
	default:
		return true
	}
}

// Room
// ========================================

// Room Goroutine loop: handles incoming publish and subscription requests
func (r *Room) Run(ready chan struct{}) {
	defer r.cleanup()

	log.Info().Msgf("Starting room '%v'", r.Name)
	close(ready)
	for {
		select {
		case <-r.ctx.Done():
			log.Debug().Msgf("'%v' received context done", r.Name)
			return
		case msg := <-r.PublishChan:
			r.MessageCount ++
			handlePublish(r, msg)
		}
	}
}

func (r *Room) Stop() {
	r.cancel()
}

func (r *Room) Publish(msgs ...interface{}) error {	

	for _, msg := range msgs {
		var messageStruct Message
		switch msg := msg.(type) {
		case Message:
			messageStruct = msg
		case string:
			messageStruct = NewMessage(msg)
		default:
			log.Error().Msgf("Attempt to publish unsupported msg '%T'", msg)
			return fmt.Errorf("unsupported publish type '%T'", msg)
		}

		// attempt to send a message
		select {
		//ensure room is not closed by checking the context.Done() channel
		case <-r.ctx.Done():
			return r.ctx.Err() 
		//if the context is not done, we can safely send a message to the room
		default: 
			r.PublishChan <- messageStruct
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

func (r *Room) NewSubscriber() (*Subscriber, error) {
	select{
	case <-r.ctx.Done():
		return nil, fmt.Errorf("room is closed")
	default:
	}
	RecvChan := make(chan Message, buffer_sub_channel) 
	sub := &Subscriber{
		Id:       	"todo",
		RecvChan: 	RecvChan,
		Room:     	r,
		isAlive:  	true,
		mu:			sync.Mutex{},
	}
	r.subscribers.Store(sub, RecvChan)	
	log.Info().Msgf("New subscriber to room '%v'", r.Name)
	return sub, nil
}

func (r *Room) Unsubscribe(sub *Subscriber) error {
	channel, exists := r.subscribers.LoadAndDelete(sub)
	if !exists {
		return fmt.Errorf("subscriber not found")
	}
	close(channel.(chan Message))
	sub.mu.Lock()
	sub.isAlive = false
	sub.mu.Unlock()
	log.Info().Msgf("Unsubscribed '%v' from room '%v'", sub.Id, r.Name)
	return nil
}

func (s *Subscriber) Recv() <-chan Message {
	return s.RecvChan
}

func (s *Subscriber) IsAlive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isAlive
}

// Internal func
// ========================================

func handlePublish(r *Room, msg Message) {
	log.Debug().Msgf("Room '%v' publishing '%v'", r.Name, msg.Payload)
	//iterate over sync.Map in r.subscribers
	wg := sync.WaitGroup{}
	r.subscribers.Range(func(sub, channel interface{}) bool {
		castSub := sub.(*Subscriber)
		subChan := channel.(chan Message)
		wg.Add(1)
		go func(){
			select {	//non blocking send
			case subChan <- msg:
			default:
				//TODO think about how we handle slow readers or dead readers
				//subscriber doesn't exist or is too slow, let's remove them from list
				r.Unsubscribe(castSub)
			}
			wg.Done()
		}()
		return true
	})
	wg.Wait()
}

func (r *Room) cleanup() {
	// remove all subscribers 
	r.subscribers.Range(func(sub, channel interface{}) bool {
		log.Debug().Msgf("Stopping subscriber '%v'", sub.(*Subscriber).Id)
		err := r.Unsubscribe(sub.(*Subscriber))
		if err != nil {
			log.Error().Msgf("Unable to unsub '%v' due to '%v'", sub.(*Subscriber).Id, err)
		}
		// subChan := channel.(chan Message)
		// close(subChan)
		return true
	})
	
	// Close all channels
	close(r.PublishChan)
	log.Info().Msgf("Cleaned and closed room '%v'", r.Name)
}
