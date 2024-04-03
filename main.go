package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Message struct {
	Topic        string
	Payload      interface{}
	CreationTime time.Time
}

type Subscriber struct {
	Channel     chan interface{}
	Unsubscribe chan bool
	SubTime     time.Time
}

type MessageBroker struct {
	Subscribers map[string][]*Subscriber // Mapping Topics to a slice of Subscriber pointers
	Mutex       sync.Mutex               // Mutex for synchronizing access to the Message Broker
}

// API
// ========================================

func NewBroker() *MessageBroker {
	return &MessageBroker{
		Subscribers: make(map[string][]*Subscriber),
		Mutex:       sync.Mutex{},
	}
}

func (broker *MessageBroker) Subscribe(topic string) *Subscriber {
	// initialize subscriber with a new channel and unsubscribe channel
	subscriber := &Subscriber{
		Channel:     make(chan interface{}),
		Unsubscribe: make(chan bool),
		SubTime:     time.Now(),
	}
	broker.Mutex.Lock()
	defer broker.Mutex.Unlock() //ensure Mutex is unlocked for the next subscribe operation
	// Add the subscriber to the broker's topic slice
	broker.Subscribers[topic] = append(broker.Subscribers[topic], subscriber)
	// Return the new subscriber
	return subscriber
}

func (broker *MessageBroker) Unsubscribe(topic string, sub *Subscriber) error {
	broker.Mutex.Lock()
	defer broker.Mutex.Unlock()
	// iterate over broker's topic map, then iterate over subscribers slice
	for i, s := range broker.Subscribers[topic] {
		if s == sub {
			// close the subscriber's channel to signal that it should stop listening
			close(sub.Channel)
			// remove subscriber from slice
			broker.Subscribers[topic] = append(broker.Subscribers[topic][:i], broker.Subscribers[topic][i+1:]...)
			return nil
		}
	}
	//return error if subscriber not found
	return fmt.Errorf("subscriber not found %v", sub)
}

func (broker *MessageBroker) Publish(topic string, payload interface{}) error {
	defer broker.Mutex.Unlock()
	broker.Mutex.Lock()
	// quick check for topic
	if _, ok := broker.Subscribers[topic]; !ok {
		return fmt.Errorf("topic not found %s", topic)
	}
	message := &Message{
		Topic:        topic,
		Payload:      payload,
		CreationTime: time.Now(),
	}

	for _, sub := range broker.Subscribers[topic] {
		sub.Channel <- *message
	}
	return nil
}

// Main
// ========================================

func main() {
	// Configure logging to write to both a file and standard output
	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open log file")
	}
	defer file.Close()

	// Create a multiwriter to write logs to both file and standard output
	multi := zerolog.MultiLevelWriter(file, zerolog.ConsoleWriter{Out: os.Stdout})

	// Set up ZeroLog with multiwriter
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()

	// Example log messages
	log.Info().Msg("Starting chat application")

}
