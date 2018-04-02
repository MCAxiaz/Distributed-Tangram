package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
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
	localRPCAddr := flag.Int("p", 0, "address to expose")

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println("usage: go run client.go [-c remote-address] [-p rpc-port] [address]")
		return
	}

	addr := flag.Args()[0]

	rand.Seed(time.Now().UTC().UnixNano())

	config, err := readConfig()
	if err != nil {
		log.Fatalln(err)
	}
	var game *tangram.Game
	if *remoteAddr == "" {
		game, err = tangram.NewGame(config, ":"+strconv.Itoa(*localRPCAddr))
	} else {
		game, err = tangram.ConnectToGame(*remoteAddr, ":"+strconv.Itoa(*localRPCAddr))
	}

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
