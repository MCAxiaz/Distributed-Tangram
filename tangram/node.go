package tangram

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"../lamport"
	"github.com/hsanjuan/go-libp2p-gorpc"
	"github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	nat "github.com/libp2p/go-libp2p-nat"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-swarm"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/multiformats/go-multiaddr"
)

// Node is the exposed RPC interface for a tangram node
type Node struct {
	game   *Game
	player *Player
	host   host.Host
	server *rpc.Server
	client *rpc.Client
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
	priv, pub, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return
	}

	pid, err := peer.IDFromPublicKey(pub)
	if err != nil {
		return
	}

	ps := peerstore.NewPeerstore()
	ps.AddPrivKey(pid, priv)
	ps.AddPubKey(pid, pub)

	maddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	if err != nil {
		return
	}

	nat := nat.DiscoverNAT()
	nat.PortMapAddrs([]multiaddr.Multiaddr{maddr})
	pubAddr, ok := nat.MappedAddrs()[maddr]
	if ok {
		log.Printf("[startNode] NAT successful: %s", pubAddr.String())
	} else {
		err = fmt.Errorf("Cannot estalibsh NAT at :%d", port)
		return
	}

	ctx := context.Background()
	network, err := swarm.NewNetwork(ctx, []multiaddr.Multiaddr{maddr}, pid, ps, nil)
	if err != nil {
		return
	}

	host := basichost.New(network)

	pubMaddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("%s/ipfs/%s", pubAddr.String(), host.ID().Pretty()))
	if err != nil {
		return
	}
	log.Printf("Address: %s", pubMaddr.String())

	player := Player{
		ID:   pid,
		Addr: pubMaddr.String(),
	}

	node = new(Node)
	node.player = &player
	node.host = host

	server := rpc.NewServer(host, "/rpc")
	server.Register(node)

	node.server = server
	node.client = rpc.NewClientWithServer(host, "/rpc", server)

	return
}

// RPC

// Connect connects to a node with the new player's information
func (node *Node) Connect(ctx context.Context, req *ConnectRequest, res *ConnectResponse) (err error) {
	fmt.Println("[Connect] Player Addr:", req.Player.Addr)
	for _, player := range node.game.state.Players {
		if req.Player.ID == player.ID {
			return fmt.Errorf("Player ID = %d is already in the game", player.ID)
		}
	}

	node.game.state.Players = append(node.game.state.Players, &req.Player)
	node.addPeer(req.Player.Addr)
	*res = ConnectResponse{node.game.GetState(), node.game.GetConfig(), node.player}
	return
}

// func (node *Node) addPeer(player *Player) (err error) {
// 	node.host.Network().Peerstore().AddAddr(player.ID, player.Addr, peerstore.PermanentAddrTTL)
// 	return
// }

func (node *Node) addPeer(addr string) (player *Player, err error) {
	ma, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return
	}

	pid, err := ma.ValueForProtocol(multiaddr.P_IPFS)
	if err != nil {
		return
	}

	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		return
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := multiaddr.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
	targetAddr := ma.Decapsulate(targetPeerAddr)

	node.host.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)

	player = &Player{
		ID:   peerid,
		Addr: ma.String(),
	}
	return
}

// GetState returns the current game state
func (node *Node) GetState(ctx context.Context, req int, res *GameState) (err error) {
	*res = *node.game.GetState()
	return
}

// GetTime returns the local timer
func (node *Node) GetTime(ctx context.Context, req int, res *time.Duration) (err error) {
	*res = node.game.GetTime()
	return
}

// LockTan locks the tan according to request
func (node *Node) LockTan(ctx context.Context, req LockTanRequest, ok *bool) (err error) {
	log.Println("[Node.LockTan]")
	*ok, err = node.game.lockTan(req.Tan, req.Player, req.Time)
	return
}

// MoveTan moves the tan according to request
func (node *Node) MoveTan(ctx context.Context, req MoveTanRequest, ok *bool) (err error) {
	log.Println("[Node.Move]")
	*ok, err = node.game.moveTan(req.Tan, req.Location, req.Rotation, req.Time)
	return
}

// Ping simply confirms that the connection is good
func (node *Node) Ping(ctx context.Context, incID PlayerID, ok *bool) (err error) {
	//do something
	*ok = true
	return
}

func (node *Node) call(dest peer.ID, svcName string, svcMethod string, args, reply interface{}) error {
	return node.client.Call(dest, svcName, svcMethod, args, reply)
}
