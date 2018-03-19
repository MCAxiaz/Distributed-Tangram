// package proj2_e3v8_e6y9a_g2u9a_j2d0b_u6x9a
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: go run client.go [address]")
		return
	}

	addr := os.Args[1]

	http.HandleFunc("/", index)
	http.HandleFunc("/ws", webSocketHandler)

	fmt.Println("Listening to requests at addr", addr)
	http.ListenAndServe(addr, nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}

func webSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		messageType, p, err := conn.ReadMessage()

		if err != nil {
			log.Println(err)
			return
		}

		// TODO: currently just echoes. We should unmarshall the message and act accordingly based on event type
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}
	}
}
