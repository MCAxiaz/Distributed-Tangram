package tangram

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// AddrPool is a struct with the following fields:
// - Pool: A map of string addresses and its latency
// - Count: How many seconds (approximately) left until we switch host
// - Votes: Current votes for next host
type AddrPool struct {
	Pool  map[string]time.Duration
	Count uint64
	Votes []*Vote
}

// Vote consists of a Voter and a Selected Host, and both are players structs
type Vote struct {
	Voter        *Player
	SelectedHost *Player
}

var addrPoolMutex = &sync.Mutex{}

// unknownLatency specifies latencies that have not been measured or is unknown
const unknownLatency = -1

// hostSwitchTimeout is a setting for how long before switching a host
const hostSwitchTimeout = uint64(60)

// addrPool is a pointer and a global variable
var addrPool = NewAddrPool()

// NewAddrPool creates a new address pool
func NewAddrPool() *AddrPool {
	return &AddrPool{
		Pool:  make(map[string]time.Duration, 0),
		Count: hostSwitchTimeout,
	}
}

// Decrement decrements the AddrPool's counter
func (a *AddrPool) Decrement() {
	a.Count--
}

// Reset will reset the countdown
func (a *AddrPool) Reset() {
	a.Count = hostSwitchTimeout
}

// Empty will empty the votes
func (a *AddrPool) Empty() {
	a.Votes = make([]*Vote, 0)
}

// CountIsPositive checks if the countdown counter is still greater than 0
func (a *AddrPool) CountIsPositive() bool {
	if a.Count < 1 {
		return true
	}

	return false
}

// AddAddressToPool adds address to the address pool
func (a *AddrPool) AddAddressToPool(addr string) {
	addrPoolMutex.Lock()
	_, ok := a.Pool[addr]
	if !ok {
		a.Pool[addr] = unknownLatency
	}
	addrPoolMutex.Unlock()
}

// UpdateLatency updates the latency of the corresponding address
func (a *AddrPool) UpdateLatency(addr string, latency time.Duration) {
	addrPoolMutex.Lock()
	a.Pool[addr] = latency
	addrPoolMutex.Unlock()
}

// selectHost will check all of the latencies collected
func (a *AddrPool) selectHost() string {
	addrPoolMutex.Lock()
	defer addrPoolMutex.Unlock()
	maxDuration := 100000 * time.Second // The connection will be dropped long before this will be ever reached
	min := maxDuration
	host := ""
	for addr, latency := range a.Pool {
		if latency < min {
			min = latency
			host = addr
		}
	}
	return host
}

// SwitchHost will allow nodes to vote for the fastest host and then
// broadcast the result to other nodes
func (a *AddrPool) SwitchHost(game *Game, players []*Player) {
	host := a.selectHost()
	var hostPlayer Player

	// Figure out the host player
	for _, player := range players {
		if player.Addr == host {
			hostPlayer = *player
		}
	}

	// Now broadcast who you think should be the host player
	for _, player := range players {
		if player.ID == game.node.player.ID {
			continue
		}

		client, err := game.pool.getConnection(player)
		if err != nil {
			log.Println(err.Error())
			err = nil
			continue
		}

		vote := a.setupVote(game.node.player, &hostPlayer)
		var ok bool
		err = client.Call("Node.RelayHost", &vote, &ok)
		if err != nil {
			fmt.Println("[Relay Host]: Cannot broadcast vote.")
			err = nil
			continue
		}
	}

	// TODO: Remove vote if voter voted more than once

	// TODO: Need to tally up votes. If there are any ties, tie break at random.

	// TODO: Once a consensus on a single host is reached, change hosts
}

// setupVote is for setting up votes by having a voter and a nomination for a host.
func (a *AddrPool) setupVote(voter *Player, hostPlayer *Player) *Vote {
	return &Vote{
		Voter:        voter,
		SelectedHost: hostPlayer,
	}
}
