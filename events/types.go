package events

// EventType is the list of possible types this application will support
type EventType int

const (
	// Subscribe events request the node register the client as a participant
	Subscribe EventType = iota
	// GetState events provide the client with the state stored at this node
	GetState
	// UpdateState events request the node update its internal state
	UpdateState
	// Unsubscribe events signal that the client is disconnecting
	Unsubscribe
)

// Event represents an event sent to this endpoint
type Event struct {
	Type EventType
}
