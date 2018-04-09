package host

// AddrPool is a struct with the following fields:
// - Pool: A map of string addresses and its latency
// - Access: The current address of the node accessing the pool
// - Time: The timestamp of the node accessing the struct
type AddrPool struct {
	Pool   map[string]int
	Access string
	Time   uint64
}

// Lock is a distributed lock that gets broadcasted over the network for permission
// to take the lock for updating AddrPool.
// Should block until lock is received successfully.
func (a *AddrPool) Lock() {
	// TODO
}

// Unlock is broadcasted throughout the network to unlock AddrPool
func (a *AddrPool) Unlock() {
	// TODO
}

// UnknownLatency specifies latencies that have not been measured or is unknown
const UnknownLatency = -1

// AddAddressToPool adds address to the address pool
func (a *AddrPool) AddAddressToPool(addr string) {
	a.Lock()
	_, ok := a.Pool[addr]
	if !ok {
		a.Pool[addr] = UnknownLatency
	}
	a.Unlock()
}

// UpdateLatency updates the latency of the corresponding address
func (a *AddrPool) UpdateLatency(addr string, latency int) {
	a.Lock()
	a.Pool[addr] = latency
	a.Unlock()
}

// SelectHost will check all of the latencies collected
func (a *AddrPool) SelectHost() string {
	a.Lock()
	defer a.Unlock()
	maxInt := int(^uint(0) >> 1)
	min := maxInt
	host := ""
	for addr, latency := range a.Pool {
		if latency < min {
			min = latency
			host = addr
		}
	}
	return host
}
