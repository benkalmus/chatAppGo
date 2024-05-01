# Chat app 

## Introduction

This app is intended to be an experimenting excercise in golang to create a message broker (pub-sub) implementation with some surrounding feature layers (listed below).

### Implementation in `pkg/pubsub/`

## Features (TODO)

- Asynchronous message passing
  - Buffered Publish and Subscribe
- Multiple Connection protocols:
  - WebSockets
  - UDP
  - to support future impl.:
    - gRPC?
- Word censor, filters out inappropriate messages
- Rate limiting. Prevents a publisher from spamming a chat
- Persistent chat
  - PostgreSQL DB store
- User Auth?
  - Could only allow users with a validated email address? 

## Development Features
- Unit Tests
- Continuous Tests (consider a framework)
- Load tests
- Perf Profiling
- Docker containerized 

# Implementation Details 

## Asynchronous Message broker

Message `Broker` consists of many `Rooms`. Each `Room` is a goroutine containing a map of `Subscriber` channels and `Publisher` channels. Messaging is handled by a `chan` which is returned on `Subscribe` to a `Room`.
Publishing is designed to be non-blocking, therefore Subscribers too slow to read messages are unsubscribed from a Room.

## WebSockets 

The app allows hooking a websocket connection to the `Subscriber` API. `Subscriber` contains `recv` methods. Most likely will use Json to communicate with front end. 

## UDP 

This should be very similar to Websockets. With a method to (de)serialise messages, consider protobuf, jsonb, or some other type that is size efficient. Agree on message protocol, e.g. subscribe,topic, publisher,topic,message. 

## Rate limiting

Can be implemented per Publisher, or per Room basis. I think room-based rate limits make more sense rather than some publishers being given more leniency as to how many messages they can spam. This could however make sense in a room that has some kind of "admin" user. Could be considered in the future with user auth. 
Websocket connection limits by IP 

## Persistent chat

Creating a DB backend for this project.
Consider methods of storing and restoring room history. 

Some rooms may have a Permanent Storage subscriber that stores data to an existing entry in a DB. When a client connects, they may subscribe to a room and also request chat history if available. This side of the project will require another design pattern. If a chat has lots of history, it's more efficient to load it in chunks via timestamps, e.g: now - 1 hour.

## User Auth

Creating a front end for this project. 
Although rate limiting can reduce malicious users spamming messages, the app is still vulnerable to loosely termed denial of service where a bad actor spawns web socket clients programmatically, allowing actors to spam many messages per each new connection. 

To prevent this, we'll need to authenticate each new connection with some kind of credentials, e.g:
  - session token & cookies 
  - User Subscription DB: username, email & password with email verification links
  - challenge-response auth. captcha. Before a WS connection can be opened, the user must satisfy the challenge.

Some of these solutions will require data collection, and hence may be encroaching EU data protection laws. 

