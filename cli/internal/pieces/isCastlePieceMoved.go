package pieces



func CastlePieceMoved(game *GameState, move string) {
	switch move {
	case "Ke1":
		game.Castle.WhiteKingMoved = true
	case "Rh1":
		game.Castle.WhiteRookKingsideMoved = true
	case "Ra1":
		game.Castle.WhiteRookQueensideMoved = true
	case "Ke8":
		game.Castle.BlackKingMoved = true
	case "Rh8":
		game.Castle.BlackRookKingsideMoved = true
	case "Ra8":
		game.Castle.BlackRookQueensideMoved = true
	}
}
