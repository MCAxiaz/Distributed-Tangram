package tangram

// Game is the public interface of a tangram game
type Game struct {
	state       *GameState
	node        *Node
	subscribers []*chan bool
}

// NewGame starts a new Game
func NewGame(config *GameConfig, localAddr string) (game *Game, err error) {
	state := new(GameState)

	node, err := startNode(localAddr)
	node.state = state
	if err != nil {
		return
	}

	state.Config = config
	state.Host = node.player
	state.Timer = 0

	state.Tans = make([]*Tan, len(state.Config.Tans))
	for i, tan := range state.Config.Tans {
		state.Tans[i] = new(Tan)
		*state.Tans[i] = *tan
	}

	state.Players = make([]*Player, 1)
	state.Players[0] = node.player

	game = new(Game)
	game.state = state
	game.node = node
	game.subscribers = make([]*chan bool, 0)
	return
}

func (game *Game) Subscribe() *chan bool {
	channel := make(chan bool, 1)
	game.subscribers = append(game.subscribers, &channel)
	return &channel
}

// GetState retrieves the current state of the board
func (game Game) GetState() *GameState {
	return game.state
}
