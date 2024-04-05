package pubsub

import (
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

// ========================================
// Types 
// ========================================
type PubSub struct {
	// set of room ptrs
	Rooms map[string]*Room
	// maybe
	// Rooms map[*Room]struct{}
	
	// access to rooms
	mutex sync.Mutex
}


type Room struct {
	Name string
	// for async publishing of messages
	Publish chan Message
	//for async buffering of sub/unsub actions
	Actions chan Subscriber 
	subscribers sync.Map // map[Subscriber]struct{}
}
// func (r1 *Room) Equals(r2 *Room) bool {
//     return r1.name == r2.name
// }
// func (r1 *Room) NotEquals(r2 *Room) bool {
//     return !r1.Equals(r2)
// }

type Subscriber struct {
	id string
}

type Message struct {
	// unique  msg hash
	id []byte
	timestamp time.Time
	Payload interface{} 
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
func NewRoom(name string) *Room {
	return &Room{
		Name: name,
		Publish: make(chan Message),
		Actions: make(chan Subscriber),
	}
}

// Room loop: handles incoming publish and subscription requests
func (r *Room) run() {
	log.Info().Msgf("Starting room %v", r.name)


}

// Actions  


// Internal func