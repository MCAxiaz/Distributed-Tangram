package tangram

import (
	"log"
	"math/rand"
	"net"
	"net/rpc"
)

// Node is the exposed RPC interface for a tangram node
type Node struct {
	state    *GameState
	player   *Player
	listener net.Listener
}

// ConnectRequest request argument for Node.Connect
type ConnectRequest struct {
	Player Player
}

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
	player.ID = rand.Uint32()
	player.Addr = addr
	return
}

// RPC

// Connect connects to a node with the new player's information
func (node *Node) Connect(req *ConnectRequest, res *int32) (err error) {
	for i, player := range node.state.Players {
		if req.Player.ID == player.ID {
			*node.state.Players[i] = req.Player
		}
	}

	return
}
