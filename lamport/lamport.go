package lamport

// Clock is an implementation of lamport clock logical time algorithm
type Clock struct {
	Counter Time
}

// Time is the logical time returned by all methods of a Clock
type Time = uint64

// Time returns the current local time for the lamport clock
func (l *Clock) Time() Time {
	return l.Counter
}

// Increment increaments the local time and returns its value after incrementing
func (l *Clock) Increment() Time {
	l.Counter++
	return l.Counter
}

// Witness updates the local time to be at least one greater than the input value
// Returns whether the local time is increased
func (l *Clock) Witness(v Time) bool {
	if v < l.Counter {
		return false
	}

	l.Counter = v
	return true
}
