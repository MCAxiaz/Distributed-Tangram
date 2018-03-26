package tangram

import (
	"fmt"
	"log"
	"net/rpc"
	"time"

	"../lamport"
)

// Game is the public interface of a tangram game
type Game struct {
	state       *GameState
	config      *GameConfig
	node        *Node
	pool        *connectionPool
	subscribers []*chan bool
}

// NewGame starts a new Game
func NewGame(config *GameConfig, localAddr string) (game *Game, err error) {
	node, err := startNode(localAddr)
	if err != nil {
		return
	}

	state := initState(config, node.player)
	// TODO Sometimes this needs to be nil to signify lack of a host
	state.Host = node.player

	game = &Game{
		state:       state,
		config:      config,
		node:        node,
		pool:        new(connectionPool),
		subscribers: make([]*chan bool, 0),
	}

	node.game = game
	return
}

func ConnectToGame(addr string, localAddr string) (game *Game, err error) {
	node, err := startNode(localAddr)
	if err != nil {
		return
	}

	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		return
	}

	var res ConnectResponse
	err = client.Call("Node.Connect", ConnectRequest{*node.player}, &res)
	if err != nil {
		return
	}

	config := res.config
	state := res.state

	game = &Game{
		state:       state,
		config:      config,
		node:        node,
		pool:        new(connectionPool),
		subscribers: make([]*chan bool, 0),
	}

	go game.syncTime(state.Players[0])
	return
}

func initState(config *GameConfig, player *Player) (state *GameState) {
	state = &GameState{
		Timer: time.Now(),
	}

	state.Tans = make([]*Tan, len(config.Tans))
	for i, tan := range config.Tans {
		state.Tans[i] = new(Tan)
		*state.Tans[i] = *tan
	}

	state.Players = make([]*Player, 1)
	state.Players[0] = player
	return
}

// Subscribe returns a channel that outputs a value when the game state is updated
func (game *Game) Subscribe() *chan bool {
	channel := make(chan bool, 1)
	game.subscribers = append(game.subscribers, &channel)
	return &channel
}

// GetState retrieves the current state of the board
func (game *Game) GetState() *GameState {
	return game.state
}

// GetTime returns the time since the game started
func (game *Game) GetTime() time.Duration {
	return time.Now().Sub(game.state.Timer)
}

func (game *Game) GetConfig() *GameConfig {
	return game.config
}

func (game *Game) syncTime(player *Player) (err error) {
	client, err := rpc.Dial("tcp", player.Addr)
	if err != nil {
		return
	}

	var d1, d2 time.Duration
	err = client.Call("Node.GetTime", 0, &d1)
	if err != nil {
		return
	}

	err = client.Call("Node.GetTime", 0, &d2)
	if err != nil {
		return
	}

	t0 := time.Now()
	rtt := d2 - d1

	newTime := t0.Add(-rtt / 2).Add(-d2)
	// TODO Add a debug flag?
	if true {
		oldTime := game.state.Timer
		d := newTime.Sub(oldTime).Nanoseconds()
		log.Printf("Time Sync with Player %d, d = %d\n", player.ID, d)
	}
	game.state.Timer = newTime

	return
}

// ObtainTan tries to gain control of the specified Tan
// This function blocks until the Tan is confirmed to be controlled
// This function is NOT guaranteed thread safe
func (game *Game) ObtainTan(id TanID) (ok bool, err error) {
	tan := game.state.getTan(id)
	if tan == nil {
		err = fmt.Errorf("[ObtainTan] Requested tan ID = %d is not found", id)
		return
	}

	time := tan.Clock.Increment()

	// Ask everyone for the tan!
	n := 0
	okChan := make(chan bool, len(game.state.Players))
	for _, player := range game.state.Players {
		client, err := game.pool.getConnection(player)
		// TODO handle error properly
		if err != nil {
			continue
		}

		go func(client *rpc.Client, player PlayerID) {
			var ok bool
			client.Call("Node.LockTan", LockTanRequest{id, player, time}, ok)
			// TODO handle error properly?
			if err != nil {
				okChan <- true
			}
			okChan <- ok
		}(client, game.node.player.ID)
		n++
	}

	// We expect n confirmations
	for ; n > 0; n-- {
		ok = <-okChan
		if !ok {
			return
		}
	}
	return
}

func (game *Game) lockTan(tanID TanID, playerID PlayerID, time lamport.Time) (ok bool, err error) {
	tan := game.state.getTan(tanID)
	if tan == nil {
		err = fmt.Errorf("[lockTan] Requested tan ID = %d is not found", tanID)
		return
	}

	// TODO we need lock around all the updates
	ok = tan.Clock.Witness(time)
	if ok {
		tan.Player = playerID
	}
	return
}
