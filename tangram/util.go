package tangram

import "math"

func (state *GameState) getTan(id TanID) *Tan {
	for _, tan := range state.Tans {
		if tan.ID == id {
			return tan
		}
	}
	return nil
}

func (state *GameState) getPlayer(id PlayerID) *Player {
	for _, player := range state.Players {
		if player.ID == id {
			return player
		}
	}
	return nil
}

func (game *Game) dropPlayer(id PlayerID) error {
	for i, player := range game.state.Players {
		if player.ID == id {
			copy(game.state.Players[i:], game.state.Players[i+1:])
			game.state.Players[len(game.state.Players)-1] = nil
			game.state.Players = game.state.Players[:len(game.state.Players)-1]
			game.pool.dropConnection(id)

			return nil
		}
	}
	return nil
}

func withinMargin(a Point, b Point, margin int32) bool {
	return math.Abs(float64(a.X-b.X)) <= float64(margin) && math.Abs(float64(a.Y-b.Y)) <= float64(margin)
}

func add(a Point, b Point) Point {
	result := a
	result.X += b.X
	result.Y += b.Y
	return result
}
