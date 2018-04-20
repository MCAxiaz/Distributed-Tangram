package tangram

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// AddrPool is a struct with the following fields:
// - Pool: A map of int player IDs and their latency
type AddrPool struct {
	Pool  map[int]time.Duration
	Mutex *sync.Mutex
}

// unknownLatency specifies latencies that have not been measured or is unknown
const unknownLatency = -1

// hostSwitchTimeout is a setting for how long before switching a host
const hostSwitchTimeout = 60

// NewAddrPool creates a new address pool
func NewAddrPool() *AddrPool {
	return &AddrPool{
		Pool:  make(map[int]time.Duration, 0),
		Mutex: &sync.Mutex{},
	}
}

// UpdateLatency updates the latency of the corresponding address
func (a *AddrPool) UpdateLatency(id int, latency time.Duration) {
	a.Mutex.Lock()
	a.Pool[id] = latency
	a.Mutex.Unlock()
}

func (game *Game) getLowestPlayerLatency() (latency int) {
	minLatency := 0
	for _, player := range game.state.Players {
		if player.AvgLatency < minLatency {
			latency = player.AvgLatency
		}
	}
	return
}

// SendLatenciesOver gets other nodes to send their average latencies over
func (game *Game) SendLatenciesOver() {
	for _, player := range game.state.Players {
		if player.ID == game.node.player.ID {
			continue
		}
		client, err := game.pool.getConnection(player)
		if err != nil {
			log.Println(err.Error())
			err = nil
			continue
		}

		var latency *int
		var args *Dict
		err = client.Call("Node.GetLatency", &args, &latency)
		if err != nil {
			fmt.Println("[Get Latency]: Cannot get latency from peer.")
			err = nil
		}
		player.AvgLatency = *latency
	}
}

// Election will start an election with nodes with higher IDs
func (game *Game) Election() {
	// Tell others to send their latencies over
	game.SendLatenciesOver()
	// If you have the highest average latency, you are the boss
	if game.node.player.AvgLatency == game.getLowestPlayerLatency() {
		game.state.Host = game.node.player
		client, err := game.pool.getConnection(game.node.player)
		if err != nil {
			log.Println(err.Error())
			err = nil
		}

		go func(game *Game) {
			var ok bool
			var err error
			err = client.Call("Node.ConnectToMe", &game.node.player, &ok)
			if err != nil {
				fmt.Println("[Connect To Me]: Cannot broadcast new host.")
				err = nil
			}
		}(game)
		return
	}

	for _, player := range game.state.Players {
		// Ignore all IDs lower and your own
		if player.ID <= game.node.player.ID {
			continue
		}

		client, err := game.pool.getConnection(player)
		if err != nil {
			log.Println(err.Error())
			err = nil
			continue
		}

		go func() {
			var ok bool
			var args *Dict
			err = client.Call("Node.HostElection", &args, &ok)
			if err != nil {
				fmt.Println("[Host Election]: Cannot broadcast host election.")
				err = nil
			}
		}()
	}
	return
}

// CalculateAvgLatency will check all of the latencies collected and
// average them.
func (game *Game) CalculateAvgLatency() (avg int) {
	game.latency.Mutex.Lock()
	defer game.latency.Mutex.Unlock()
	avg = 0
	sum := 0
	count := 0
	for _, latency := range game.latency.Pool {
		sum += int(latency)
		count++
	}
	if count != 0 {
		avg = sum / count
	}
	return avg
}
