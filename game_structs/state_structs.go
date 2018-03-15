package game_structs

// GameState is a struct that holds the state of a game.
// It has the following fields:
// - Tan: It holds a list of tans ordered and prioritized by last interaction.
// - Timer: The timer counts down.
// - Players: It holds the players currently in the game.
// - Shape: It is the silhouette of the shape players are trying to form with tans.
// - Host: The player that is hosting the game.
type GameState struct {
	Tan     *[]Tan
	Timer   uint32
	Players *[]Player
	Shape   *Silhouette
	Host    *Player
}

// Tan is a struct that holds the following information:
// - Shape: The shape of the tan
// - Player: The player that last held the tan
// - Location: The location of the tan on a canvas
// - Rotation: Alignment of tan in increments of 5 degrees
type Tan struct {
	Shape    *Shape
	Player   *Player // A fixed tan from a shape silhouette would have a nil player value
	Location *Point  // Location is a pointer to the Point struct that holds the x and y coordinates of a tan's location by its centre.
	Rotation uint32
}

// Shape contains information to create an SVG string.
// Shapes are hardcoded and do not change throughout the game.
// - Points: eg. if we have a rectangle, the list of points would include the four corners
// - and the coordinate of the points would be based on the fact that the centre of the shape is (0, 0).
type Shape struct {
	Points *[]Point // Points using centre point of shape (location field of Tan) as origin.
	Edges  *[]Line
	Fill   string
	Stroke string
}

// Line is a struct that holds information about a line that joins from point (x1, y1) to (x2, y2).
// Coordinates are relative to the centre of the shape.
type Line struct {
	Point1 *Point
	Point2 *Point
}

// Point is a struct containing a pair of x and y coordinates.
type Point struct {
	X int32
	Y int32
}

// Player is a struct that holds player information.
type Player struct {
	ID   uint32
	Name string
}

// Silhouette is a struct holding information on the silhouette of the target shape.
type Silhouette struct {
	FixedTan *[]Tan
}
