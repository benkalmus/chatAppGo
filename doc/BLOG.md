# Detailing my journey in creating a Chat App in Go

Before I start writing a Chat App application, I want to nail some foundations that will ease in designing and building this application as it grows in complexity. Firstly, let's setup a project directory structure where we'll place our files, I want to organise my packages, tests, docs and other non-source code files logically and find them their own places. Here's my current dir hierarchy 

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

Naturally, we want to add source control, for this we'll initialize a new git repo with git init. 
Next, let's reduce repetitive tasks by offsetting them to a Makefile, where we can define commands to run a formatter, build and run our application and tests. This file will inevitably grow as we add more functionality to our project, For example, we will add a dockerfile and we'll most certainly want to automate build and deployment. 

go mod init 

