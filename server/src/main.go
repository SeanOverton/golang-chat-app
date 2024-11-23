package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
	"context"
	"encoding/json"
	"os"
)

// converts http connections to websockets
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

// in memory arrays for tracking clients and messages
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

var rabbitMQURL = os.Getenv("RABBITMQ_URL")

var q amqp.Queue
var ch *amqp.Channel

func main() {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		panic("Failed to connect to RabbitMQ")
	}
	defer conn.Close()
	
	ch, err = conn.Channel()
	if err != nil {
		panic("Failed to open a channel")
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"chat",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)

	q, err = ch.QueueDeclare(
		"", // name
		false,   // durable
		false,   // delete when unused
		true,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		panic("Failed to create queue")
	}
	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		"chat", // exchange
		false,
		nil,
	)

	// http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", handleConnections)
   
	go handleMessages()
	go handleMessagesFromQueue()
   
	fmt.Println("Server started on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Error starting server: " + err.Error())
	}

}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	clients[conn] = true

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println(err)
			delete(clients, conn)
			return
		}

		broadcast <- msg
	}
}

func publishToExchange(msg Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Print(msg)

	marshalled_obj, err := json.Marshal(&msg)
	if err != nil {
		panic("Failed to json marshal msg")
	}
	// example := "hello world"
	err = ch.PublishWithContext(ctx,
		"chat",     // exchange
		"", // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing {
			ContentType: "application/json",
			Body:        marshalled_obj,
	})

	if err != nil {
		panic("Failed to publish message to queue")
	}
}

func handleMessages() {
	for {
		msg := <-broadcast

		publishToExchange(msg)

		// for client := range clients {
		// 	err := client.WriteJSON(msg)
		// 	if err != nil {
		// 		fmt.Println(err)
		// 		client.Close()
		// 		delete(clients, client)
		// 	}
		// }
	}
}

func handleMessagesFromQueue() {
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		panic("Failed to create queue consumer")
	}

	for {
		msg := <-msgs

		rcv_msg := Message{}
		err = json.Unmarshal(msg.Body, &rcv_msg)

		fmt.Printf("%s", rcv_msg)

		for client := range clients {
			err := client.WriteJSON(rcv_msg)
			if err != nil {
				fmt.Println(err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}