package tangram

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
	"time"
)

// AddrPool is a struct with the following fields:
// - Pool: A map of int player IDs and their latency
type AddrPool struct {
	MyPing  map[PlayerID]time.Duration
	AvgPing map[PlayerID]time.Duration
	Mutex   *sync.Mutex
}

// unknownLatency specifies latencies that have not been measured or is unknown
const unknownLatency = -1

// hostSwitchTimeout is a setting for how long before switching a host
const hostSwitchTimeout = 60

// NewAddrPool creates a new address pool
func NewAddrPool() *AddrPool {
	return &AddrPool{
		MyPing:  make(map[PlayerID]time.Duration),
		AvgPing: make(map[PlayerID]time.Duration),
		Mutex:   &sync.Mutex{},
	}
}

// UpdateLatency updates the latency of the corresponding address
func (a *AddrPool) UpdateLatency(id PlayerID, latency time.Duration) {
	a.Mutex.Lock()
	log.Printf("[UpdateLatency] ID = %d, latency = %d", id, latency)
	a.MyPing[id] = latency
	a.Mutex.Unlock()
}

// SendLatenciesOver gets other nodes to send their average latencies over
func (game *Game) SendLatenciesOver() {
	var wg sync.WaitGroup
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

		wg.Add(1)
		go func(client *rpc.Client, player PlayerID) {
			defer wg.Done()
			var latency time.Duration
			err := client.Call("Node.GetLatency", 0, &latency)
			if err != nil {
				fmt.Println("[Get Latency] Cannot get latency from %d.", player)
				return
			}
			log.Printf("[Get Latency] Got latency from %d.", player)
			game.latency.AvgPing[player] = latency
		}(client, player.ID)
	}
	wg.Wait()
}

// Election will start an election with nodes with higher IDs
func (game *Game) Election() {
	// Don't allow any action during an election
	game.lock.Lock()
	defer game.lock.Unlock()
	game.latency.Mutex.Lock()
	defer game.latency.Mutex.Unlock()

	// Tell others to send their latencies over
	// TODO This should also signal that an election started
	// Until the election ends, nobody should be allowed to update their latency
	game.SendLatenciesOver()

	myPlayer := game.node.player.ID
	myLatency := game.GetAvgLatency()

	// We use Bully Algorithm to figure out who should be the new host
	// Let's first see if someone else would be the host
	for player, latency := range game.latency.AvgPing {
		if less(player, latency, myPlayer, myLatency) {
			player := game.state.getPlayer(player)
			if player == nil {
				continue
			}

			client, err := game.pool.getConnection(player)
			if err != nil {
				continue
			}

			var ok bool
			err = client.Call("Node.HostElection", 0, &ok)
			if err != nil {
				continue
			}

			// If we reached here, someone finished the election?
			// TODO Figure out what it means to reach here
			log.Printf("Node.HostElection returned from ID = %d", player.ID)
			return
		}
	}

	// It is now our turn to become host
	// TODO tell everyone to listen to you
	game.state.Host = myPlayer
	log.Printf("[Election] Declaring host ID = %d", myPlayer)
	for _, player := range game.interestingPlayers() {

		if player.ID == myPlayer {
			continue
		}

		client, err := game.pool.getConnection(player)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		var ok bool
		err = client.Call("Node.ConnectToMe", myPlayer, &ok)
		if err != nil {
			log.Println(err.Error())
			continue
		}
	}

	return
}

// GetAvgLatency will check all of the latencies collected and
// average them.
func (game *Game) GetAvgLatency() (avg time.Duration) {
	avg = 0
	var sum time.Duration
	var count time.Duration
	for _, latency := range game.latency.MyPing {
		sum += latency
		count++
	}
	if count != 0 {
		avg = sum / count
	}
	return avg
}

func less(player1 PlayerID, latency1 time.Duration, player2 PlayerID, latency2 time.Duration) bool {
	return (latency1 < latency2) || (latency1 == latency2 && player1 < player2)
}
