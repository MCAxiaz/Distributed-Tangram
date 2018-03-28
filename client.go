// package proj2_e3v8_e6y9a_g2u9a_j2d0b_u6x9a
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"./tangram"
	"./webserver"
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

	config, err := readConfig()
	if err != nil {
		log.Fatalln(err)
	}
	// config.Size = tangram.Point{X: 800, Y: 800}
	// config.Tans = make([]*tangram.Tan, 0)
	// config.Target = make([]*tangram.Tan, 0)
	game, err := tangram.NewGame(config, ":0")

	if err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc("/ws", getWebSocketHandler(game))
	http.Handle("/", http.FileServer(http.Dir("web")))

	fmt.Println("Listening to requests at addr", addr)
	err = http.ListenAndServe(addr, nil)

	if err != nil {
		log.Fatalln(err)
	}
}

func readConfig() (config *tangram.GameConfig, err error) {
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return
	}

	config = new(tangram.GameConfig)
	err = json.Unmarshal(file, config)
	return
}

func getWebSocketHandler(game *tangram.Game) func(http.ResponseWriter, *http.Request) {
	handler := webserver.NewHandler(game)

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatalln(err)
		}

		err = handler.Handle(conn)
		return
	}
}
