package tangram

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// AddrPool is a struct with the following fields:
// - Pool: A map of int player IDs and their latency
// - Count: How many seconds (approximately) left until we switch host
// - Votes: Current votes for next host
type AddrPool struct {
	Pool  map[int]time.Duration
	Wait  time.Time
	Votes []*Vote
	Mutex *sync.Mutex
}

// Vote consists of a Voter and a Selected Host, and both are players structs
type Vote struct {
	Voter        *Player
	SelectedHost *Player
}

// unknownLatency specifies latencies that have not been measured or is unknown
const unknownLatency = -1

// hostSwitchTimeout is a setting for how long before switching a host
const hostSwitchTimeout = 60

// NewAddrPool creates a new address pool
func NewAddrPool() *AddrPool {
	return &AddrPool{
		Pool:  make(map[int]time.Duration, 0),
		Wait:  time.Now(),
		Mutex: &sync.Mutex{},
	}
}

// Empty will empty the votes
func (a *AddrPool) Empty() {
	a.Mutex.Lock()
	a.Votes = make([]*Vote, 0)
	a.Mutex.Unlock()
}

// CheckTime checks if the countdown counter is still greater than 0
func (a *AddrPool) CheckTime() bool {
	if a.Wait.Sub(time.Now()) >= (hostSwitchTimeout * time.Second) {
		return true
	}

	return false
}

// UpdateLatency updates the latency of the corresponding address
func (a *AddrPool) UpdateLatency(id int, latency time.Duration) {
	a.Mutex.Lock()
	a.Pool[id] = latency
	a.Mutex.Unlock()
}

// selectHost will check all of the latencies collected and
// return the ID of the host with the lowest latency
func (game *Game) selectHost() int {
	game.latency.Mutex.Lock()
	defer game.latency.Mutex.Unlock()
	maxDuration := time.Duration(1<<63 - 1)
	min := maxDuration
	// By default, the host address will be the current player's
	hostID := game.node.player.ID
	for id, latency := range game.latency.Pool {
		if latency < min {
			min = latency
			hostID = id
		}
	}
	return hostID
}

// SwitchHost will allow nodes to vote for the fastest host and then
// broadcast the result to other nodes
func (game *Game) SwitchHost() {
	players := game.state.Players
	host := game.selectHost()
	var hostPlayer *Player

	// Figure out the host player
	for _, player := range players {
		if player.ID == host {
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

		vote := game.latency.setupVote(game.node.player, hostPlayer)
		go func(vote *Vote) {
			var ok bool
			err = client.Call("Node.RelayHost", &vote, &ok)
			if err != nil {
				fmt.Println("[Relay Host]: Cannot broadcast vote.")
				err = nil
			}
		}(vote)
	}

	// Wait until number of votes >= number of different players excluding yourself
	for {
		if len(game.latency.Votes) >= len(game.state.Players)-1 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Need to tally up votes. If there are any ties, nominated candidate with higher player ID becomes host
	hostPlayer = game.tallyVotes()
	game.latency.Empty()

	// Once a consensus on a single host is reached, change hosts
	game.state.Host = hostPlayer
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

			go func(game *Game) {
				var ok bool
				err = client.Call("Node.ConnectToNewHost", &game.node.player, &ok)
				if err != nil {
					fmt.Println("[Connect New Host]: Cannot broadcast new host.")
					err = nil
				}
			}(game)
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
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	for _, voteInList := range a.Votes {
		if (*voteInList).Voter.ID == vote.Voter.ID {
			return
		}
	}
	a.Votes = append(a.Votes, vote)
}

func (game *Game) tallyVotes() (hostPlayer *Player) {
	var runningCount = make(map[PlayerID]int)
	var id PlayerID
	for _, vote := range game.latency.Votes {
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
		} else if count == maxCount {
			// Tiebreaker done with higher player ID
			if playerID > hostPlayerID {
				hostPlayerID = playerID
			}
		}
	}

	hostPlayer = game.state.getPlayer(hostPlayerID)

	return hostPlayer
}
