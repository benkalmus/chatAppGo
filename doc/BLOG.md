# Detailing my journey in creating a Chat App in Go

Before I start writing a Chat App application, I want to nail some foundations that will ease in designing and building this application as it grows in complexity. Firstly, let's setup a project directory structure where I'll place my files, I want to organise my packages, tests, docs and other non-source code files logically and find them their own places. Here's my current dir hierarchy 

```sh
$ tree
├── bin
│   └── chat_app
├── cmd
│   └── main.go
├── doc
│   ├── assets
│   └── BLOG.md
├── docker
│   └── Dockerfile
├── internal
│   └── message_broker
│       └── message_broker.go
├── logs
│   └── app.log
├── pkg
│   └── pub_sub
│       └── pub_sub.go
├── README.md
├── CHANGELOG.md
├── vsn.mk
├── Makefile
├── go.mod
├── go.sum
```

Naturally, I want to add source control, for this I'll initialize a new git repo with git init. 
Next, let's reduce repetitive tasks by offsetting them to a Makefile, where I can define commands to run a formatter, build and run our application and tests. This file will inevitably grow as I add more functionality to our project, For example, I will add a dockerfile and I'll most certainly want to automate build and deployment. 

go mod init 


## Setting up a logger

Most important part of any application is its ability to tell us what's going on internally, and for this I'll need a well configured logger that will allow us to troubleshoot and debug. From my research, zerolog was a highly recommended logging package, due to its high performance, structure, multi-level writer (io and file) and customizability. Even though I'd like to minimize the number of dependencies I wish to use for this project, I consider logging important enough to justify adding an external dependency to this project.

## Shoulders of giants

For my chat application I decided to go with a common design pattern, Publisher/Subscriber. It's a type of event driven architecture that allows us to implement a concurrent message passing mechanism. Before I get started writing my own pub sub from scratch, I'd like to look up some open source implementations in the Golang community as a starting point, as well as a benchmark comparison against my own implementation. Aside from Int. and Cont. testing I'm also planning on writing load test and profiling my application. This will satisfy my inner longing for extreme performance. 
I often enjoy looking up others' projects and implementations of design patterns, not only does it give a great starting point but more importantly it engages the inner Sherlock Holmes, to try and figure out how someone else's project works and the inner  architect in me, to contemplate their decisions as well as criticize their solution (only so that I can come up with a better one). I think it provides an insight into how others approach solutions, but more importantly it's a learning experience as reading and understanding someone else's code is not as easy as it first appears. This is a useful skill which I strongly encourage my readers to try, as it is often the case that one implementation is not strictly better than another, but rather it comes with its own pros and tradeoffs. 

That aside, I decided to look up a simple pub sub implementation, I've tidied up the code and added some comments. As this is a piece of code I won't be sharing, it belongs in the internal/ dir.

## Design considerations

As this is a chat application intended to be used on my website, I will want to optimize my solution for some criteria and not others, such as I want Publishing of messages to be as fast as possible. Naturally, as this is written in Go, I will want to come up with an asynchronous solution, while keeping the code simple to maintain and read by others. I'm going to assume that the reader of this blog has a solid understanding of CSP concurrency model in Go, as well as goroutines, channels and mutexes. 
Since in my chat app, each chat room will be independed of another, we can spawn a goroutine per room and treat it completely independently. I will still want to track each room so that a connecting client can choose to join a set of rooms, for this I will use a set of Room pointers. 

I want my implementation to be non-blocking, the key point being that I can write new layers upon the pub sub and allow implementations to process other work in the background. For example, I may wish to have a DB subscriber process, responsible for storing messages permanently, I would not want to block this process from writing SQL requests just because it's waiting for a message to be published, or waiting to be subscribed to a new chat room that was just created. 

For this reason, I decided to go create a Publishing and an Actions channel for each Room. The publish channel, as the name implies, is the entrypoint to queue and buffer messages. Since I want to prioritise publishing of messages, I wouldn't want to block the room everytime a new subscription is made or existing client unsubscribes. We expect users to come and go, but not as frequently as the messages themselves. Because a subscription action will need to lock access to the Subscribers list/slice or set, (undecided on the data structure to use just yet, it's likely that I'll experiment with various types), this will cause the room to stop publishing messages. This is not ideal, it means that anytime a new user joins a room, we have to stop what we're doing and stop sending messages until that user is subscribed. Instead, we're going to assign a channel to subscribe and unsubscribe actions, and the room will process these messages when it's free. This means that Subscribe and Unsubscribe to our pub sub will be asynchronous and as such the Subscriber implementation will need to account for an acknoledgement from the room that they're subscribed and ready to receive for messages. We can do this by responding with a Channel when it's ready. 

## Writing a Pub Sub

TODO, describe the process of writing the code, with some explanations of types, pointers, structs and interfaces. 

API decisions
Check if Room is running? 
What happens if a Room crashes? 


# Testing

Okay, so I've built a PubSub containing a set of Rooms, each Room containing a set of Subscribers, and each Subscriber containing an individual link to its room and a channel for receiving messages. I deliberately left Message type to be an interface{}, as this will allow me to change  what kind of messages i want to send depending on the project. The idea is that I can reuse this code in the future and create a pub sub for an entirely different purpose to chat app. 

## Unit testing 

Now that I'm happy with my initial iteration, I want to start testing this code, ensuring that what I've written is robust and will cover various scenarios and use cases that one might encounter when building a pub sub. I will want to attempt to break my code, and patch any failing tests. This will also highlight any issues with my design, such as where have I forgotten to return errors, to handling edge cases.


### test ideas 
- Add pubsub
    - add room 
        - add sub
        - remove sub
        - clean up 
        - subscriber that crashes
        - 
        - publish 
            - send message, multiple messages 
            - custom message payload types
            - messages arrive in FIFO order, check via timestamp 

    - add many rooms

    - remove room 
        - add sub to a non existing room 
        - remove sub from a non existing room 
        - publish "

- remove pub sub instance 
    - cleans up all rooms, subs, 


- remove pubsub
    - removes all rooms and unsubscribes all users gracefully 


## Load testing 

This is where the good stuff comes in. I will want to observe how my application behaves as I increase the number of concurrent pubsubs, rooms, subscribers, publishers and messages. I should come  upwith a list of parameters to tweak and test them against various scenarios. 
Params:
How does the pub sub  perf scale when we increase:
    - Num rooms? 
    - Num subs?
    - Num publishers? 
    - Num Messages?
    - Message Size? 
    - Concurrent Subscribers?
    - Concurrent Unsubscribers?
    - Concurrent publishers? 
    - Buffer sizes

I'm mostly concerned with how my pubsub scales rather than the actual values. 

All of these will have varying impacts on the application performance and I will want to find my strengths and weaknesses, and see if we can improve the weakness without too many trade offs.

Metrics I want to measure
- Throughput (messages published per sec)
- Latencies (Time between message publish and message received)
- CPU & core Usage (am I utilizing all cores available to me?)
- Memory (memory leaks, huge memory use, can I get away with passing pointers instead of making copies?)

google some other metrics I can measure


## Performance Profiling

So I've identified some problems, let's run a profiler on our load tests and see where I can make optimizations. 


# Features

Now that I'm happy with my pub sub and how it works, it's time to make it "do stuff". 

## WebSockets