// package proj2_e3v8_e6y9a_g2u9a_j2d0b_u6x9a
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"./tangram"
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
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatalln(err)
		}

		for {
			_, _, err := conn.ReadMessage()

			if err != nil {
				log.Println(err)
				return
			}

			state := game.GetState()
			svg := render(state)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(svg)); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func render(state *tangram.GameState) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf(`<svg width="%d" height="%d">`, state.Config.Size.X, state.Config.Size.Y))
	for _, tan := range state.Tans {
		buf.WriteString(renderTan(tan))
	}
	buf.WriteString(`</svg>`)
	return buf.String()
}

func renderTan(tan *tangram.Tan) string {
	transform := fmt.Sprintf(`translate(%d, %d) rotate(%d)`, tan.Location.X, tan.Location.Y, tan.Rotation)

	var buf bytes.Buffer
	for i, point := range tan.Shape.Points {
		command := "L"
		if i == 0 {
			command = "M"
		}

		buf.WriteString(fmt.Sprintf("%s %d %d ", command, point.X, point.Y))
	}
	buf.WriteString("Z")

	d := buf.String()
	path := fmt.Sprintf(`<path fill="%s" stroke="%s" transform="%s" d="%s">`, tan.Shape.Fill, tan.Shape.Stroke, transform, d)
	return path
}
