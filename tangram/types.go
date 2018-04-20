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
	Tans    []*Tan `json:"tans"`
	Timer   time.Time
	Players []*Player
	Host    PlayerID `json:"host"`
	Solved  bool
}

// GameConfig is the starting configuration of a game
// - Tans: Tans position when the game begins
// - Target: The shape players are trying to form with tans.
type GameConfig struct {
	Size    Point
	Offset  Point
	Margin  int32
	Tans    []*Tan
	Targets []*TargetTan `json:"targets"`
	Host    bool
}

// Tan is a struct that holds the following information:
// - ID: The ID of the tan
// - Shape: The shape of the tan
// - Player: The ID of the player controlling the tan
// - Location: The location of the tan on a canvas
// - Rotation: Alignment of tan in increments of 5 degrees
// - Clock: A logical clock for this tan
type Tan struct {
	ID        TanID         `json:"id"`
	Shape     *Shape        `json:"shape"`
	ShapeType ShapeType     `json:"type"`
	Player    PlayerID      `json:"player"`
	Location  Point         `json:"location"`
	Rotation  Rotation      `json:"rotation"`
	Clock     lamport.Clock `json:"clock"`
	Matched   bool
}

// Tan is a struct that holds the following information:
// - ID: The ID of the tan
// - Shape: The shape of the tan
// - Player: The ID of the player controlling the tan
// - Location: The location of the tan on a canvas
// - Rotation: Alignment of tan in increments of 5 degrees
// - Clock: A logical clock for this tan
type TargetTan struct {
	Shape     *Shape    `json:"shape"`
	ShapeType ShapeType `json:"type"`
	Location  Point     `json:"location"`
	Rotation  Rotation  `json:"rotation"`
}

// Shape contains information to create an SVG string.
// Shapes are hardcoded and do not change throughout the game.
// - Points: eg. if we have a rectangle, the list of points would include the four corners
// - and the coordinate of the points would be based on the fact that the centre of the shape is (0, 0).
// - The points are ordered in a clockwise fashion.
type Shape struct {
	Points []Point `json:"points"` // Points using centre point of shape (location field of Tan) as origin.
	Fill   string  `json:"fill"`
	Stroke string  `json:"stroke"`
}

// Point is a struct containing a pair of x and y coordinates.
type Point struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
}

// Player is a struct that holds player information.
type Player struct {
	ID   PlayerID
	Name string
	Addr string
}

// PlayerID is the ID of a Player
// A valid ID must be non-negative
type PlayerID = int

// NoPlayer is the PlayerID of an uncontrolled tan
const NoPlayer PlayerID = -1

// ShapeType is the type of a Tan or Solution Tan

type ShapeType string

const (
	LTri  ShapeType = "LTri"
	MTri            = "MTri" // Matches a Mid-sized Triangle or two Small Triangles
	STri            = "STri"
	Cube            = "Cube"  // Matches a Cube or two Small Triangles
	Pgram           = "Pgram" // Matches a Parallelogram or two Small Triangles

)

// TanID is the ID of a Tan
type TanID = uint32

// Rotation is the type of rotation in degrees
type Rotation = uint32
