# Chat app 

## Introduction

This app is intended to be a simple message broker (pub-sub) implementation in golang.


## Features

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

Message `Broker` consists of many `Rooms`. Each `Room` is a process containing a map of `Subscriber` channels and `Publisher` channel.
`Publisher` channel is used to asynchronously receive `Message`s, which are stored in-memory by the `Room` until it is free to publish downstream. A `room` may be busy during a `subscribe` or `unsubscribe` operation, in which case it will lock the `Subscriber` slice until it is done. Messages during that time are buffered and processed when unlocked. 

`Subscriptions` are done async and buffered, as to be processed later because, in this implementation we wish Publishing to take higher priority. If lots of messages are being published, a timer is created to force the processing of subscriptions.

TODO:
- keep the implementation as simple as possible, do not worry about multiple room subscribers, that can be added as a wrapper for the existing API. such as, MultiSub containg a slice of []Subscriber instances.
- do not worry about message data type, but add useful metadata like Timestamps

## WebSockets 

The app allows hooking a websocket connection to the `Subscriber` API. `Subscriber` contains `recv` methods. Most likely will use Json to communicate with front end. 

## UDP 

This should be very similar to Websockets. With a method to (de)serialise messages, consider protobuf, jsonb, or some other type that is size efficient. Agree on message protocol, e.g. subscribe,topic, publisher,topic,message. 

## Rate limiting

Can be implemented per Publisher, or per Room basis. I think room-based rate limits make more sense rather than some publishers being given more leniency as to how many messages they can spam. This could however make sense in a room that has some kind of "admin" user. Could be considered in the future with user auth. 
Websocket connection limits by IP 

## Persistent chat

Creating a backend for our project.
Consider methods of storing and restoring room history. KV or SQL? 

Some rooms may have a Permanent Storage subscriber which stores data to an existing entry in a DB. When a client connects, they may subscribe to a room and also request chat history if available. This side of the project will require a completely different design pattern to the pub-sub mechanism. If a chat has lots of history, it's more efficient to load it based on timestamps, e.g: now - 1 hour.

## User Auth

Creating a front end for our project. 
Although rate limiting can reduce malicious users spamming messages, the app is still vulnerable to loosely termed DDoS attacks where a bad actor spawns web socket clients programmatically, allowing actors to spam many more messages per each new connection. 

To prevent this, we'll need to authenticate each new connection with some kind of credentials:
  - session token & cookies 
  - full on Users DB: username, email & password
  - challenge-response auth. captcha. Before a WS connection can be opened, the user must satisfy the challenge.

Some of these solutions will require data collection, and hence may be encroaching EU data protection laws. 

