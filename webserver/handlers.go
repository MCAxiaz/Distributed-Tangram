package webserver

import (
	"encoding/json"
	"fmt"
	"log"

	"../tangram"
	"github.com/gorilla/websocket"
)

type Message struct {
	MsgType string `json:"type"`
}

type OutputMessage struct {
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
			log.Printf("[Handle] Inbound Message %s", msg)
			if err != nil {
				close(msgChan)
				return
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
			err = handler.handleMessage(conn, msg)
			if err != nil {
				log.Printf("[Handle] Error: %s", err.Error())
			}
		}
	}
}

func (handler *Handler) handleChange(conn *websocket.Conn) {
	state := handler.game.GetState()
	conn.WriteJSON(OutputMessage{"state", state})
}

func (handler *Handler) handleMessage(conn *websocket.Conn, data []byte) (err error) {
	var msg Message
	err = json.Unmarshal(data, &msg)
	if err != nil {
		return
	}

	switch msg.MsgType {
	case "GetState":
		err = handler.handleGetState(conn, data)
	case "ObtainTan":
		err = handler.handleObtainTan(conn, data)
	default:
		err = fmt.Errorf("Unsupported Message %s", msg.MsgType)
	}
	return
}

func (handler *Handler) handleGetState(conn *websocket.Conn, data []byte) (err error) {
	state := handler.game.GetState()
	err = conn.WriteJSON(OutputMessage{"state", state})
	return
}

type ObtainTanMessage struct {
	Tan tangram.TanID `json:"tan"`
}

func (handler *Handler) handleObtainTan(conn *websocket.Conn, data []byte) (err error) {
	var msg ObtainTanMessage
	err = json.Unmarshal(data, &msg)
	if err != nil {
		return
	}
	ok, err := handler.game.ObtainTan(msg.Tan)
	if err != nil {
		return
	}
	// Do something with it?
	log.Println(ok)
	return
}

func handleError(err error) {
	if err != nil {
		log.Println(err)
	}
}
