package chess


func IsaCastlingMove(move string)bool  {
	moves := map[string]bool{
		"O-O":   true,
		"O-O-O": true,
		"e1g1":  true,
		"e1c1":  true,
		"e8g8":  true,
		"e8c8":  true,
	}
	return moves[move]
}