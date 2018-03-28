package tangram

import (
	"net/rpc"
)

type connectionPool struct {
	connections map[PlayerID]*rpc.Client
}

func newConnectionPool() *connectionPool {
	return &connectionPool{
		connections: make(map[PlayerID]*rpc.Client),
	}
}

func (pool *connectionPool) getConnection(player *Player) (client *rpc.Client, err error) {
	client, ok := pool.connections[player.ID]
	if ok {
		return
	}

	client, err = pool.connect(player.Addr)
	if err != nil {
		return
	}

	pool.connections[player.ID] = client
	return
}

func (pool *connectionPool) connect(addr string) (client *rpc.Client, err error) {
	client, err = rpc.Dial("tcp", addr)
	return
}
