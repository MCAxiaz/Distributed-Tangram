package webserver

import (
	"encoding/json"
	"log"

	"../tangram"
	"github.com/gorilla/websocket"
)

type Message struct {
	MsgType string      `json:"type"`
	Data    interface{} `json:"data"`
}

type Handler struct {
	game *tangram.Game
}

func NewHandler(game *tangram.Game) *Handler {
	return &Handler{game}
}

func (handler *Handler) Handle(conn *websocket.Conn) (err error) {
	changeChan := handler.game.Subscribe()
	defer handler.game.Unsubscribe(changeChan)

	msgChan := make(chan []byte, 10)
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				close(msgChan)
			}
			msgChan <- msg
		}
	}()

	for {
		select {
		// Handle change
		case _, ok := <-changeChan:
			if !ok {
				log.Println("[Handle] Change Channel closed")
				return
			}
			handler.handleChange(conn)
		// Handle msg
		case msg, ok := <-msgChan:
			if !ok {
				log.Println("[Handle] Message Channel closed")
				return
			}
			handler.handleMessage(conn, msg)
		}
	}
}

func (handler *Handler) handleChange(conn *websocket.Conn) {
	state := handler.game.GetState()
	conn.WriteJSON(Message{"state", state})
}

func (handler *Handler) handleMessage(conn *websocket.Conn, data []byte) {
	var msg Message

	json.Unmarshal(data, &msg)
	switch msg.MsgType {
	case "GetState":
		state := handler.game.GetState()
		conn.WriteJSON(Message{"state", state})
	default:
		log.Println(msg.Data)
	}
}

func handleError(err error) {
	if err != nil {
		log.Println(err)
	}
}
