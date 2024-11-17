# Go chat CLI

I made a CLI chat app in Golang because I was bored and wanted to practice my GoLang skills

# Run rabbitMQ

`docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:4.0-management`

# Run the server via docker

## Build Image

`cd server`
`docker build --platform=linux/amd64 -t golang-chat-server .`

## Run container

`docker run -it --rm -p 8080:8080 golang-chat-server`

# Run the chat client as a CLI tool

`cd client/src`
`go run main.go`
