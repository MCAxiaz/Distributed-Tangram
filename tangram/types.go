package tangram

import (
	"time"

	"../lamport"
)

// GameState is a struct that holds the state of a game.
// It has the following fields:
// - Tans: It holds a list of tans ordered and prioritized by last interaction.
// - Timer: The time when the game started.
// - Players: It holds the players currently in the game.
// - Host: The player that is hosting the game.
type GameState struct {
	Tans    []*Tan
	Timer   time.Time
	Players []*Player
	Host    *Player
	Config  *GameConfig
}

// GameConfig is the starting configuration of a game
// - Tans: Tans position when the game begins
// - Target: The shape players are trying to form with tans.
type GameConfig struct {
	Size   Point
	Tans   []*Tan
	Target []*Tan
}

// Tan is a struct that holds the following information:
// - ID: The ID of the tan
// - Shape: The shape of the tan
// - Player: The ID of the player controlling the tan
// - Location: The location of the tan on a canvas
// - Rotation: Alignment of tan in increments of 5 degrees
// - Clock: A logical clock for this tan
type Tan struct {
	ID       TanID
	Shape    *Shape
	Player   PlayerID
	Location Point
	Rotation uint32
	Clock    lamport.Clock
}

// Shape contains information to create an SVG string.
// Shapes are hardcoded and do not change throughout the game.
// - Points: eg. if we have a rectangle, the list of points would include the four corners
// - and the coordinate of the points would be based on the fact that the centre of the shape is (0, 0).
// - The points are ordered in a clockwise fashion.
type Shape struct {
	Points []Point // Points using centre point of shape (location field of Tan) as origin.
	Fill   string
	Stroke string
}

// Point is a struct containing a pair of x and y coordinates.
type Point struct {
	X int32
	Y int32
}

// Player is a struct that holds player information.
type Player struct {
	ID   PlayerID
	Name string
	Addr string
}

// PlayerID is the ID of a Player
// A valid ID must be non-negative
// An ID of -1 means nil
type PlayerID = int

// TanID is the ID of a Tan
type TanID = uint32
