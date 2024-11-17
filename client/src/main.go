package main

import (
	"bufio"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
	"github.com/gorilla/websocket"
	"encoding/json"
	"strings"
	"fmt"
)

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// connect to server
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	// Initiate user input reader
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("What is your username today: ")
	username, err := reader.ReadString('\n')
	username = strings.TrimSpace(username)
		
	log.Println("***Welcome to the chat! ***")
	fmt.Printf("[%s]: ", username)
	
	// listen for messages and print to terminal
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			rcv_msg := Message{}
			err = json.Unmarshal([]byte(message), &rcv_msg)

			if (rcv_msg.Username != username) {
				fmt.Printf("\n[%s]: %s", rcv_msg.Username, rcv_msg.Message)
			}else{
				fmt.Printf("\n[%s(YOU)]: %s", rcv_msg.Username, rcv_msg.Message)
			}	
		}
	}()

	chat := make(chan string)

	go func() {
		for {
			// Call the reader to read user's input
			chat_message, err := reader.ReadString('\n')
			chat_message = strings.TrimSpace(chat_message)
			if err != nil {
				panic(err)
			}

			msg_obj := Message{
				Username: username,
				Message: chat_message,
			}
			marshalled_obj, _ := json.Marshal(&msg_obj)
			chat <- string(marshalled_obj)
		}
	}()

	for {
		select {
		case <-done:
			return
		case t := <-chat:
			err := c.WriteMessage(websocket.TextMessage, []byte(t))
			if err != nil {
				log.Println("Error:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}