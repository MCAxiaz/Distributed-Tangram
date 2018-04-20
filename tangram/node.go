package tangram

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-nat"
	"github.com/multiformats/go-multiaddr"

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
func startNode(addr string, playerID int) (node *Node, err error) {
	port := strings.Split(addr, ":")[1]
	ip := addr[:len(addr)-len(port)-1]

	if ip == "" {
		ip = "0.0.0.0"
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return
	}

	externalAddr := addr
	if ip != "127.0.0.1" {
		// Setup port forwarding
		externalIP, externalPort, err := mapPortLibp2p(ip, portNum)
		if err != nil {
			log.Println(err)
			log.Printf("UPnP port forwarding failed")
		} else {
			externalAddr = fmt.Sprintf("%s:%d", externalIP, externalPort)
		}
	} else {
		log.Println("Skipping port forwarding")
	}

	resolvedAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return
	}

	inbound, err := net.ListenTCP("tcp", resolvedAddr)
	if err != nil {
		return
	}

	node = new(Node)
	node.listener = inbound

	node.player = newPlayer(externalAddr, playerID)

	server := rpc.NewServer()
	server.Register(node)
	go server.Accept(inbound)
	log.Printf("Listening on %s as %d\n", addr, node.player.ID)
	return
}

func newPlayer(addr string, id int) (player *Player) {
	player = new(Player)

	if id == 0 {
		// Randomize if no id specified
		player.ID = rand.Int()
	} else {
		player.ID = id
	}

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

// ConnectToNewHost connects to new host
func (node *Node) ConnectToNewHost(host *Player, ok *bool) (err error) {

	// TODO: Unsubscribe and close connection to previous host

	client, err := rpc.Dial("tcp", host.Addr)
	if err != nil {
		return
	}

	var res ConnectResponse
	err = client.Call("Node.Connect", ConnectRequest{*node.player}, &res)
	if err != nil {
		return
	}

	fmt.Println("Connected to new host: ", host.Addr)
	*ok = true
	return
}

// HostElection receives
func (node *Node) HostElection(args *Dict, ok *bool) (err error) {
	node.game.Election()
	*ok = true
	return
}

func mapPortLibp2p(ip string, port int) (externalIP string, externalPort int, err error) {
	nat := nat.DiscoverNAT()
	if nat == nil {
		log.Println("[lib2p2] DiscoverNAT failed")
		err = fmt.Errorf("DiscoverNAT failed")
		return
	}

	addr := fmt.Sprintf("/ip4/%s/tcp/%d", ip, port)
	ma, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return
	}

	nat.PortMapAddrs([]multiaddr.Multiaddr{ma})
	externalAddr, ok := nat.MappedAddrs()[ma]
	if !ok {
		err = fmt.Errorf("Cannot find mapped address %s", ma.String())
	}

	externalIP, err = externalAddr.ValueForProtocol(multiaddr.P_IP4)
	if err != nil {
		return
	}

	externalPortStr, err := externalAddr.ValueForProtocol(multiaddr.P_TCP)
	if err != nil {
		return
	}

	externalPort, err = strconv.Atoi(externalPortStr)
	if err != nil {
		return
	}

	log.Printf("[mapPortLibp2p] Port forwarding from %s:%d to %s:%d", externalIP, externalPort, ip, port)
	return
}
