package tangram

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"strings"
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

// ConnectResponse is response argument for Node.Connect
type ConnectResponse struct {
	State  *GameState
	Config *GameConfig
	Player *Player
}

// LockTanRequest is request argument for Node.LockTan
type LockTanRequest struct {
	Tan    TanID
	Player PlayerID
	Time   lamport.Time
}

// MoveTanRequest is request argument for Node.MoveTan
type MoveTanRequest struct {
	Tan      TanID
	Location Point
	Rotation Rotation
	Time     lamport.Time
}

// startNode instantiates the RPC server which will allow for communication between client nodes
func startNode(port int) (node *Node, err error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return
	}

	inbound, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}

	node = new(Node)
	node.listener = inbound

	// Find the outbound IP address to listen to
	res, err := http.Get("https://wtfismyip.com/text")
	if err != nil {
		log.Fatalln(err)
	}

	defer res.Body.Close()
	ipBytes, err := ioutil.ReadAll(res.Body)
	ip := strings.TrimSpace(string(ipBytes[:len(ipBytes)]))

	fmt.Println(ip)

	if err != nil {
		log.Fatalln(err)
	}

	localAddr := fmt.Sprintf("%v:%v", ip, port)
	node.player = newPlayer(localAddr)

	server := rpc.NewServer()
	server.Register(node)
	go server.Accept(inbound)
	log.Printf("Listening on %s as %d\n", localAddr, node.player.ID)
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
func (node *Node) Connect(req *ConnectRequest, res *ConnectResponse) (err error) {
	for _, player := range node.game.state.Players {
		if req.Player.ID == player.ID {
			return fmt.Errorf("Player ID = %d is already in the game", player.ID)
		}
	}

	node.game.state.Players = append(node.game.state.Players, &req.Player)

	*res = ConnectResponse{node.game.GetState(), node.game.GetConfig(), node.player}
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
	log.Println("[Node.LockTan]")
	*ok, err = node.game.lockTan(req.Tan, req.Player, req.Time)
	return
}

// MoveTan moves the tan according to request
func (node *Node) MoveTan(req MoveTanRequest, ok *bool) (err error) {
	log.Println("[Node.Move]")
	*ok, err = node.game.moveTan(req.Tan, req.Location, req.Rotation, req.Time)
	return
}

// Ping simply confirms that the connection is good
func (node *Node) Ping(incID PlayerID, ok *bool) (err error) {
	//do something
	*ok = true
	return
}
