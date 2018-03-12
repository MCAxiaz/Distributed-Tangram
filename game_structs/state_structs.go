package game_structs

// GameState is a struct that holds the state of a game.
// It has the following fields:
// - Tan: It holds a list of tans ordered and prioritized by last interaction.
// - Timer: The timer counts down.
// - Players: It holds the players currently in the game.
// - Shape: It is the silhouette of the shape players are trying to form with tans.
type GameState struct {
	Tan     *[]Tan
	Timer   uint8
	Players *[]Player
	Shape   *Silhouette
	Host    *Player
}

// Tan is a struct that holds the following information:
// - Player: The player that last held the tan
// - Location: The location of the tan on a canvas
// - Rotation: Alignment of tan in increments of 5 degrees
type Tan struct {
	Player   *Player // A fixed tan would have a nil player value
	Location *Location
	Rotation uint8
}

type Axis int

// Location is a struct that holds the x and y coordinates of a tan's location by its centre.
type Location struct {
	X int
	Y int
}

// Player is a struct that holds player information.
type Player struct {
	ID   uint8
	Name string
}

// Silhouette is a struct holding information on the silhouette of the target shape.
type Silhouette struct {
	FixedTans *[]Tan
}
