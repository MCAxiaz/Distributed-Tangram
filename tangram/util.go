package tangram

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
