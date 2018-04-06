// package proj2_e3v8_e6y9a_g2u9a_j2d0b_u6x9a
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"./tangram"
	"./webserver"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	remoteAddr := flag.String("c", "", "remote client to connect to")
	rpcPort := flag.Int("p", 9000, "address to expose")

	flag.Parse()

	rand.Seed(time.Now().UTC().UnixNano())

	config, err := readConfig()
	if err != nil {
		log.Fatalln(err)
	}

	var game *tangram.Game
	if *remoteAddr == "" {
		game, err = tangram.NewGame(config, *rpcPort)
	} else {
		game, err = tangram.ConnectToGame(*remoteAddr, *rpcPort)
	}

	if err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc("/ws", getWebSocketHandler(game))
	http.Handle("/", http.FileServer(http.Dir("web")))

	var addr string
	if len(flag.Args()) == 1 {
		addr = flag.Args()[0]
		fmt.Println("Listening to requests at addr", addr)
	} else if len(flag.Args()) == 0 {
		addr = ":8080"
		fmt.Println("[Default] Listening to requests at addr", addr)
	} else {
		fmt.Println("usage: go run client.go [-c remote-address] [-p rpc-port] [address]")
		return
	}

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
