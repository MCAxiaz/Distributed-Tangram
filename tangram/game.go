package tangram

import (
	"fmt"
	"log"
	"math"
	"net/rpc"
	"sync"
	"time"

	"../lamport"
)

// Game is the public interface of a tangram game
type Game struct {
	lock        sync.RWMutex
	state       *GameState
	config      *GameConfig
	node        *Node
	pool        *connectionPool
	subscribers []chan bool
	latency     *AddrPool
}

// NewGame starts a new Game
func NewGame(config *GameConfig, addr string, playerID int) (game *Game, err error) {
	node, err := startNode(addr, playerID)
	if err != nil {
		return
	}

	state := initState(config, node.player)
	// TODO Sometimes this needs to be nil to signify lack of a host
	state.Host = node.player.ID

	game = &Game{
		state:       state,
		config:      config,
		node:        node,
		latency:     NewAddrPool(),
		pool:        newConnectionPool(),
		subscribers: make([]chan bool, 0),
	}

	node.game = game

	go game.heartbeat()

	return
}

// ConnectToGame connects to an existing game at addr
func ConnectToGame(remoteAddr string, addr string, playerID int) (game *Game, err error) {
	node, err := startNode(addr, playerID)
	if err != nil {
		return
	}

	client, err := rpc.Dial("tcp", remoteAddr)
	if err != nil {
		return
	}

	game = &Game{
		node:        node,
		latency:     NewAddrPool(),
		pool:        newConnectionPool(),
		subscribers: make([]chan bool, 0),
	}
	node.game = game

	game.lock.Lock()

	var res ConnectResponse
	err = client.Call("Node.Connect", ConnectRequest{*node.player}, &res)
	if err != nil {
		return
	}

	config := res.Config
	state := initState(config, node.player)

	game.state = state
	game.config = config

	game.lock.Unlock()

	game.witnessState(res.State)
	game.syncTime(state.getPlayer(res.Player.ID))

	go game.heartbeat()

	return
}

func (game *Game) heartbeat() {
	for {
		for _, player := range game.interestingPlayers() {
			if player.ID == game.GetPlayer().ID {
				continue
			}

			client, err := game.pool.getConnection(player)
			if err != nil {
				log.Println(err.Error())
				game.dropPlayer(player.ID)
				delete(game.latency.MyPing, player.ID)
				continue
			}

			go func(player *Player, client *rpc.Client) {
				start := time.Now()
				err := game.pingPlayer(player.ID, client)
				end := time.Now()
				elapsed := end.Sub(start)

				if err != nil {
					// If there is a disconnection with the host
					if game.state.Host == player.ID {
						game.Election()
					}
					game.dropPlayer(player.ID)
					return
				}

				game.latency.UpdateLatency(player.ID, elapsed)
			}(player, client)
		}
		time.Sleep(5 * time.Second)
	}
}

func (game *Game) pingPlayer(id PlayerID, client *rpc.Client) (err error) {
	var ok bool
	err = client.Call("Node.Ping", game.GetPlayer().ID, &ok)
	return
}

func (game *Game) connectToPeer(player *Player) (err error) {
	client, err := rpc.Dial("tcp", player.Addr)
	if err != nil {
		fmt.Println("connectToPeer error")
		return
	}

	var res ConnectResponse
	err = client.Call("Node.Connect", ConnectRequest{*game.GetPlayer()}, &res)
	if err != nil {
		return
	}
	game.witnessState(res.State)

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

// Returns the gamestate with solved true if solved, false otherwise.
func checkSolution(config *GameConfig, state *GameState) {
	numMatched := 0
	tanMap := make(map[ShapeType][]int)

	//first sort tans into types
	for i, tan := range state.Tans {
		tanMap[tan.ShapeType] = append(tanMap[tan.ShapeType], i)
		tan.Matched = false
	}

	// Match based on ShapeType
	for _, target := range config.Targets {
		switch target.ShapeType {
		// case MTri:
		case Cube:
			numMatched += matchMultiple(state, config, tanMap[target.ShapeType], target, 90.0)
		case Pgram:
			numMatched += matchMultiple(state, config, tanMap[target.ShapeType], target, 180.0)
		default:
			numMatched += matchMultiple(state, config, tanMap[target.ShapeType], target, 360.0)
		}
	}
	if numMatched == len(config.Targets) {
		state.Solved = true
	} else {
		state.Solved = false
	}
}

// returns 1 if matched, 0 otherwise.
// mod allows shapes like square to match to multiple angles. 360 default
func matchMultiple(state *GameState, config *GameConfig, indexes []int, target *TargetTan, mod float64) int {
	for _, index := range indexes {
		if isMatch(config, state.Tans[index], target, mod) {
			state.Tans[index].Matched = true
			return 1
		}
	}
	return 0
}

func isMatch(config *GameConfig, tan *Tan, target *TargetTan, mod float64) bool {
	rotationMatches := math.Mod(float64(tan.Rotation), mod) == math.Mod(float64(target.Rotation), mod)
	return withinMargin(add(target.Location, config.Offset), tan.Location, config.Margin) && rotationMatches
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
	if game.state.Host == game.GetPlayer().ID {
		state := copyState(game.state)
		for _, player := range game.interestingPlayers() {
			if player.ID == game.GetPlayer().ID {
				continue
			}

			client, err := game.pool.getConnection(player)
			if err != nil {
				log.Println(err.Error())
				continue
			}

			go func(client *rpc.Client) {
				var ok bool
				err := client.Call("Node.PushUpdate", state, &ok)
				if err != nil {
					log.Println(err.Error())
				}
			}(client)
		}
	}
	for _, sub := range game.subscribers {
		select {
		case sub <- true:
		default:
		}
	}
	checkSolution(game.config, game.state)
}

// GetState retrieves the current state of the board
func (game *Game) GetState() *GameState {
	game.lock.RLock()
	stateCopy := copyState(game.state)
	game.lock.RUnlock()
	return stateCopy
}

// GetTime returns the time since the game started
func (game *Game) GetTime() time.Duration {
	game.lock.RLock()
	t := time.Now().Sub(game.state.Timer)
	game.lock.RUnlock()
	return t
}

// GetConfig returns the config of the game
func (game *Game) GetConfig() *GameConfig {
	return game.config
}

func (game *Game) GetPlayer() *Player {
	return game.node.player
}

func (game *Game) syncTime(player *Player) (err error) {
	log.Printf("[syncTime] Start with player %d", player.ID)
	client, err := rpc.Dial("tcp", player.Addr)
	if err != nil {
		return
	}

	var d1, d2 time.Duration
	err = client.Call("Node.GetTime", 0, &d1)
	if err != nil {
		return
	}
	log.Printf("[syncTime] Got first response")

	err = client.Call("Node.GetTime", 0, &d2)
	if err != nil {
		return
	}
	log.Printf("[syncTime] Got second response")

	t0 := time.Now()
	rtt := d2 - d1

	newTime := t0.Add(-rtt / 2).Add(-d2)
	game.lock.Lock()
	if true {
		oldTime := game.state.Timer
		d := newTime.Sub(oldTime).Nanoseconds()
		log.Printf("Time Sync with Player %d, d = %d\n", player.ID, d)
	}
	game.state.Timer = newTime
	game.lock.Unlock()

	return
}

// ObtainTan tries to gain control of the specified Tan
// This function blocks until the Tan is confirmed to be controlled
// This function is NOT guaranteed thread safe
func (game *Game) ObtainTan(id TanID, release bool) (ok bool, err error) {
	log.Printf("[ObtainTan] ID = %d\n", id)
	game.lock.Lock()
	tan := game.state.getTan(id)
	if tan == nil {
		err = fmt.Errorf("[ObtainTan] Requested tan ID = %d is not found", id)
		game.lock.Unlock()
		return
	}

	if tan.Player != NoPlayer && tan.Player != game.GetPlayer().ID {
		log.Printf("[ObtainTan] Obtaining TanID = %d failed. Already controlled by %d", id, tan.Player)
		game.lock.Unlock()
		return false, nil
	}

	playerID := game.GetPlayer().ID
	if release {
		playerID = NoPlayer
	}

	time := tan.Clock.Increment()
	game.lock.Unlock()

	// Ask everyone for the tan!
	n := 0
	okChan := make(chan bool, len(game.state.Players))
	for _, player := range game.interestingPlayers() {
		if player.ID == game.GetPlayer().ID {
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

	game.lock.Lock()
	tan.Player = playerID
	game.lock.Unlock()
	game.notify()
	return
}

// MoveTan changes the location of a Tan
// MoveTan does not block and broadcasts the content asynchronously
func (game *Game) MoveTan(id TanID, location Point, rotation Rotation) (ok bool, err error) {
	// log.Printf("[MoveTan] ID = %d\n", id)
	game.lock.Lock()
	tan := game.state.getTan(id)
	if tan == nil {
		err = fmt.Errorf("[ObtainTan] Requested tan ID = %d is not found", id)
		game.lock.Unlock()
		return
	}

	if tan.Player != game.GetPlayer().ID {
		ok = false
		game.lock.Unlock()
		return
	}

	time := tan.Clock.Increment()
	tan.Location = location
	tan.Rotation = rotation
	ok = true
	game.lock.Unlock()

	// Let everyone know!
	for _, player := range game.interestingPlayers() {
		if player.ID == game.GetPlayer().ID {
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
	game.lock.Lock()
	defer game.lock.Unlock()
	tan := game.state.getTan(tanID)
	if tan == nil {
		err = fmt.Errorf("[lockTan] Requested tan ID = %d is not found", tanID)
		return
	}

	oldTime := tan.Clock.Time()
	ok = tan.Clock.Witness(time)
	if ok {
		tan.Player = determineOwner(tan.Player, oldTime, playerID, time)

		if tan.Player != playerID {
			ok = false
		}
	}

	game.notify()
	return
}

func determineOwner(currentHolder PlayerID, tanTime lamport.Time, playerID PlayerID, newTime lamport.Time) (lockHolder PlayerID) {
	// If two requests for a tan occur at the same time, handle deterministically
	// We will use PlayerID to determine who locks the tan
	if currentHolder != NoPlayer && tanTime == newTime {
		var lesserID, greaterID PlayerID
		log.Printf("[lockTan] Resolving conflict between players %v | %v at time: %v\n", currentHolder, playerID, tanTime)

		if currentHolder < playerID {
			lesserID = currentHolder
			greaterID = playerID
		} else {
			lesserID = playerID
			greaterID = currentHolder
		}

		if tanTime%2 == 0 {
			lockHolder = lesserID
		} else {
			lockHolder = greaterID
		}

		log.Printf("[lockTan] Resolution: %v holds the lock\n", lockHolder)
	} else {
		lockHolder = playerID
	}

	return
}

func (game *Game) moveTan(tanID TanID, location Point, rotation Rotation, time lamport.Time) (ok bool, err error) {
	game.lock.Lock()
	defer game.lock.Unlock()
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

	time := newTan.Clock.Time()
	oldTime := tan.Clock.Time()
	ok := tan.Clock.Witness(time)
	log.Printf("[witnessTan] Witness ID = %d, ok = %t\n", tan.ID, ok)
	if ok {
		tan.Location = newTan.Location
		tan.Rotation = newTan.Rotation
		tan.Player = determineOwner(tan.Player, oldTime, newTan.Player, time)
	}
	checkSolution(game.config, game.state)
}

func (game *Game) witnessState(state *GameState) {
	game.state.Host = state.Host
	for _, tan := range state.Tans {
		game.witnessTan(tan)
	}
	for _, player := range state.Players {
		if game.state.getPlayer(player.ID) != nil {
			continue
		}

		log.Printf("[witnessState] Adding Player %d at %s", player.ID, player.Addr)
		game.state.Players = append(game.state.Players, player)

		if game.isPlayerInteresting(player) {
			game.connectToPeer(player)
		}
		go game.measureLatency(player)

	}

	checkSolution(game.config, state)
}

func (game *Game) interestingPlayers() []*Player {
	host := game.state.Host
	// Decentralized
	if !game.hosted() {
		return game.state.Players
	}
	// I am host, I am responsible for updating all peers
	if host == game.GetPlayer().ID {
		return game.state.Players
	}
	// I am subscribing to a host, I talk to the host alone
	hostPlayer := game.state.getPlayer(host)
	if hostPlayer != nil {
		return []*Player{hostPlayer}
	}
	return []*Player{}
}

func (game *Game) isPlayerInteresting(player *Player) bool {
	if !game.hosted() {
		return true
	}
	if game.state.Host == game.GetPlayer().ID {
		return true
	}
	if game.state.Host == player.ID {
		return true
	}
	return false
}

func (game *Game) hosted() bool {
	return game.state.Host != NoPlayer
}

func (game *Game) measureLatency(player *Player) (err error) {
	client, err := game.pool.getConnection(player)
	if err != nil {
		return
	}
	start := time.Now()
	err = game.pingPlayer(player.ID, client)
	end := time.Now()
	elapsed := end.Sub(start)
	game.latency.UpdateLatency(player.ID, elapsed)
	return
}
