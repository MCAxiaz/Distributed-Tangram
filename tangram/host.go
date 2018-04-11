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
	addrPoolMutex.Lock()
	defer addrPoolMutex.Unlock()
	return &AddrPool{
		Pool:  make(map[string]time.Duration, 0),
		Count: hostSwitchTimeout,
	}
}

// Decrement decrements the AddrPool's counter
func (a *AddrPool) Decrement() {
	addrPoolMutex.Lock()
	a.Count--
	addrPoolMutex.Unlock()
}

// Reset will reset the countdown
func (a *AddrPool) Reset() {
	addrPoolMutex.Lock()
	a.Count = hostSwitchTimeout
	addrPoolMutex.Unlock()
}

// Empty will empty the votes
func (a *AddrPool) Empty() {
	addrPoolMutex.Lock()
	a.Votes = make([]*Vote, 0)
	addrPoolMutex.Unlock()
}

// CountIsPositive checks if the countdown counter is still greater than 0
func (a *AddrPool) CountIsPositive() bool {
	if a.Count < 1 {
		return true
	}

	return false
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
	var hostPlayer *Player

	// Figure out the host player
	for _, player := range players {
		if player.Addr == host {
			hostPlayer = player
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

		vote := a.setupVote(game.node.player, hostPlayer)
		var ok bool
		err = client.Call("Node.RelayHost", &vote, &ok)
		if err != nil {
			fmt.Println("[Relay Host]: Cannot broadcast vote.")
			err = nil
			continue
		}
	}

	// Wait until number of votes == number of different players excluding yourself
	for {
		if len(a.Votes) == len(players)-1 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Need to tally up votes. If there are any ties, first nominated host
	// in the map gets to be the host
	hostPlayer = a.tallyVotes(players)
	a.Empty()

	// TODO: Once a consensus on a single host is reached, change hosts
	if hostPlayer.ID == game.node.player.ID {
		// Ask everyone to connect to you
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

			var ok bool
			err = client.Call("Node.ConnectToNewHost", &game.node.player, &ok)
			if err != nil {
				fmt.Println("[Connect New Host]: Cannot broadcast new host.")
				err = nil
				continue
			}
		}
	}
}

// setupVote is for setting up votes by having a voter and a nomination for a host.
func (a *AddrPool) setupVote(voter *Player, hostPlayer *Player) *Vote {
	return &Vote{
		Voter:        voter,
		SelectedHost: hostPlayer,
	}
}

// AddVote adds a vote to the list of votes
// It does not append the vote if the voter has already voted.
func AddVote(a *AddrPool, vote *Vote) {
	addrPoolMutex.Lock()
	defer addrPoolMutex.Unlock()
	for _, voteInList := range a.Votes {
		if (*voteInList).Voter.ID == vote.Voter.ID {
			return
		}
	}
	a.Votes = append(a.Votes, vote)
}

func (a *AddrPool) tallyVotes(players []*Player) (hostPlayer *Player) {
	var runningCount = make(map[PlayerID]int)
	var id PlayerID
	for _, vote := range a.Votes {
		id = (*vote).SelectedHost.ID
		runningCount[id]++
	}

	minVote := 0
	maxCount := minVote
	hostPlayerID := -1
	for playerID, count := range runningCount {
		if count > maxCount {
			maxCount = count
			hostPlayerID = playerID
		}
	}

	for _, player := range players {
		if hostPlayerID == player.ID {
			hostPlayer = player
			break
		}
	}

	return hostPlayer
}
