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
	subscribers []chan bool
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
		pool:        newConnectionPool(),
		subscribers: make([]chan bool, 0),
	}

	node.game = game
	return
}

// ConnectToGame connects to an existing game at addr
func ConnectToGame(addr string, localAddr string) (game *Game, err error) {
	var (
		node   *Node
		client *rpc.Client
		res    ConnectResponse
	)

	node, err = startNode(localAddr)
	if err != nil {
		return
	}

	client, err = rpc.Dial("tcp", addr)
	if err != nil {
		return
	}

	req := ConnectRequest{*node.player}

	log.Printf("[ConnectToGame] Attempting connection on %s\n", addr)
	if err = client.Call("Node.Connect", req, &res); err != nil {
		return
	}

	config := res.Config
	state := initState(config, node.player)

	attempted := make(map[string]bool)
	attempted[addr] = true

	todo := []*Player{}

	for _, player := range state.Players {
		attempted[player.Addr] = false
		todo = append(todo, player)
	}

	for {
		last := len(todo) - 1

		if last == -1 {
			break
		}

		player := todo[last]

		if !(player.Addr == node.player.Addr) && !attempted[player.Addr] {
			client, err = rpc.Dial("tcp", player.Addr)
			if err != nil {
				return
			}

			// TODO: handle some graceful way to recover from failing connections
			// The remote node shouldn't include disconnected clients in game state
			log.Printf("[ConnectToGame] Attempting connection on %s\n", player.Addr)
			if err = client.Call("Node.Connect", req, &res); err != nil {
				return
			}

			attempted[player.Addr] = true

			// Discover addresses we may not already know
			for _, other := range res.State.Players {
				if _, ok := attempted[other.Addr]; !ok {
					log.Printf("Appending undiscovered client %s\n", other.Addr)
					todo = append(todo, other)
				}
			}
		}

		todo = append(todo[:last], todo[last+1:]...)
	}

	game = &Game{
		state:       state,
		config:      config,
		node:        node,
		pool:        newConnectionPool(),
		subscribers: make([]chan bool, 0),
	}

	node.game = game

	game.witnessState(res.State)
	game.syncTime(state.getPlayer(res.Player.ID))
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
		state.Tans[i].Player = NoPlayer
	}

	state.Players = make([]*Player, 1)
	state.Players[0] = player
	return
}

// Subscribe returns a channel that outputs a value when the game state is updated
func (game *Game) Subscribe() chan bool {
	channel := make(chan bool, 1)
	game.subscribers = append(game.subscribers, channel)
	return channel
}

// Unsubscribe takes a channel reutrned by Subscribe() and remove & close it
func (game *Game) Unsubscribe(s chan bool) {
	index := -1
	for i, subscriber := range game.subscribers {
		if subscriber == s {
			close(subscriber)
			index = i
			break
		}
	}

	if index < 0 {
		panic("Channel not found")
	}

	game.subscribers[index] = game.subscribers[len(game.subscribers)-1]
	game.subscribers = game.subscribers[:len(game.subscribers)-1]
}

func (game *Game) notify() {
	for _, sub := range game.subscribers {
		select {
		case sub <- true:
		default:
		}
	}
}

// GetState retrieves the current state of the board
func (game *Game) GetState() *GameState {
	return game.state
}

// GetTime returns the time since the game started
func (game *Game) GetTime() time.Duration {
	return time.Now().Sub(game.state.Timer)
}

// GetConfig returns the config of the game
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
func (game *Game) ObtainTan(id TanID, release bool) (ok bool, err error) {
	log.Printf("[ObtainTan] ID = %d\n", id)
	tan := game.state.getTan(id)
	if tan == nil {
		err = fmt.Errorf("[ObtainTan] Requested tan ID = %d is not found", id)
		return
	}

	playerID := game.node.player.ID
	if release {
		playerID = NoPlayer
	}

	time := tan.Clock.Increment()

	// Ask everyone for the tan!
	n := 0
	okChan := make(chan bool, len(game.state.Players))
	for _, player := range game.state.Players {
		if player.ID == game.node.player.ID {
			continue
		}

		client, err := game.pool.getConnection(player)
		// TODO handle error properly
		if err != nil {
			continue
		}

		go func(client *rpc.Client, player PlayerID) {
			var ok bool
			client.Call("Node.LockTan", LockTanRequest{id, player, time}, &ok)
			// TODO handle error properly?
			if err != nil {
				log.Println(err.Error())
				okChan <- true
			}
			okChan <- ok
		}(client, playerID)
		n++
	}
	log.Printf("[ObtainTan] ID = %d. %d peer responses expected\n", id, n)

	// We expect n confirmations
	ok = true
	for n > 0 {
		ok = <-okChan
		n--
		log.Printf("[ObtainTan] ID = %d. Got response %t. %d more responses expected\n", id, ok, n)
		if !ok {
			return
		}
	}

	tan.Player = playerID
	game.notify()
	return
}

// MoveTan changes the location of a Tan
// MoveTan does not block and broadcasts the content asynchronously
func (game *Game) MoveTan(id TanID, location Point, rotation Rotation) (ok bool, err error) {
	// log.Printf("[MoveTan] ID = %d\n", id)
	tan := game.state.getTan(id)
	if tan == nil {
		err = fmt.Errorf("[ObtainTan] Requested tan ID = %d is not found", id)
		return
	}

	time := tan.Clock.Increment()
	tan.Location = location
	tan.Rotation = rotation
	ok = true

	// Let everyone know!
	for _, player := range game.state.Players {
		if player.ID == game.node.player.ID {
			continue
		}

		client, err := game.pool.getConnection(player)
		// TODO handle error properly
		if err != nil {
			log.Println(err.Error())
			continue
		}

		go func(client *rpc.Client) {
			var ok bool
			client.Call("Node.MoveTan", MoveTanRequest{id, location, rotation, time}, &ok)
		}(client)
	}

	game.notify()
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

	game.notify()
	return
}

func (game *Game) moveTan(tanID TanID, location Point, rotation Rotation, time lamport.Time) (ok bool, err error) {
	tan := game.state.getTan(tanID)
	if tan == nil {
		err = fmt.Errorf("[moveTan] Requested tan ID = %d is not found", tanID)
		return
	}

	ok = tan.Clock.Witness(time)
	if ok {
		tan.Location = location
		tan.Rotation = rotation
	}

	game.notify()
	return
}

func (game *Game) witnessTan(newTan *Tan) {
	tan := game.state.getTan(newTan.ID)
	if tan == nil {
		log.Printf("[witnessTan] Witnessed ghost ID = %d\n", newTan.ID)
		return
	}

	ok := tan.Clock.Witness(newTan.Clock.Time())
	log.Printf("[witnessTan] Witness ID = %d, ok = %t\n", tan.ID, ok)
	if ok {
		tan.Location = newTan.Location
		tan.Rotation = newTan.Rotation
		tan.Player = newTan.Player
	}
}

func (game *Game) witnessState(state *GameState) {
	for _, tan := range state.Tans {
		game.witnessTan(tan)
	}
	for _, player := range state.Players {
		if game.state.getPlayer(player.ID) != nil {
			continue
		}
		log.Printf("[witnessState] Adding Player %d", player.ID)
		game.state.Players = append(game.state.Players, player)
	}
}
