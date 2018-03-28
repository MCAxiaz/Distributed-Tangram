package tangram

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"time"

	"../lamport"
)

// Node is the exposed RPC interface for a tangram node
type Node struct {
	game     *Game
	player   *Player
	listener net.Listener
}

// ConnectRequest is request argument for Node.Connect
type ConnectRequest struct {
	Player Player
}

type ConnectResponse struct {
	state  *GameState
	config *GameConfig
}

// LockTanRequest is request argument for Node.Connect
type LockTanRequest struct {
	Tan    TanID
	Player PlayerID
	Time   lamport.Time
}

// startNode instantiates the RPC server which will allow for communication between client nodes
func startNode(localAddr string) (node *Node, err error) {
	addr, err := net.ResolveTCPAddr("tcp", localAddr)
	if err != nil {
		return
	}

	inbound, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}

	node = new(Node)
	node.listener = inbound

	localAddr = inbound.Addr().String()
	node.player = newPlayer(localAddr)

	server := rpc.NewServer()
	server.Register(node)
	go server.Accept(inbound)
	log.Println("Listening on ", localAddr)
	return
}

func newPlayer(addr string) (player *Player) {
	player = new(Player)
	player.ID = rand.Int()
	player.Addr = addr
	return
}

// RPC

// Connect connects to a node with the new player's information
func (node *Node) Connect(req *ConnectRequest, res *GameConfig) (err error) {
	for _, player := range node.game.state.Players {
		if req.Player.ID == player.ID {
			return fmt.Errorf("Player ID = %d is already in the game", player.ID)
		}
	}

	node.game.state.Players = append(node.game.state.Players, &req.Player)

	*res = *node.game.config
	return
}

// GetState returns the current game state
func (node *Node) GetState(req int, res *GameState) (err error) {
	*res = *node.game.GetState()
	return
}

// GetTime returns the local timer
func (node *Node) GetTime(req int, res *time.Duration) (err error) {
	*res = node.game.GetTime()
	return
}

// LockTan locks the tan according to request
func (node *Node) LockTan(req LockTanRequest, ok *bool) (err error) {
	*ok, err = node.game.lockTan(req.Tan, req.Player, req.Time)
	return
}
