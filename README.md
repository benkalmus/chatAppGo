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

## Development Features
- Unit Tests
- Continuous Tests (consider a framework)
- Load tests
- Perf Profiling
-  

## Solution 

