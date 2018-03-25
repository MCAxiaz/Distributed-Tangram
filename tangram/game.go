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
	node        *Node
	subscribers []*chan bool
}

// NewGame starts a new Game
func NewGame(config *GameConfig, localAddr string) (game *Game, err error) {
	node, err := startNode(localAddr)
	if err != nil {
		return
	}

	state := new(GameState)
	state.Config = config
	state.Host = node.player
	state.Timer = time.Now()

	state.Tans = make([]*Tan, len(state.Config.Tans))
	for i, tan := range state.Config.Tans {
		state.Tans[i] = new(Tan)
		*state.Tans[i] = *tan
	}

	state.Players = make([]*Player, 1)
	state.Players[0] = node.player

	game = new(Game)
	game.state = state
	game.node = node
	game.subscribers = make([]*chan bool, 0)

	node.game = game
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

	newTime := t0.Add(-rtt).Add(-d2)
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
func (game *Game) ObtainTan(id TanID) (err error) {
	// TODO Ask all nodes/host for the tan
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
