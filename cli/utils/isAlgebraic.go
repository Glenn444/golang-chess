package utils


func IsAlgebraic(move string)bool{
	//move - e2e4
	mapBoard := make(map[string]bool)
	mapBoard = map[string]bool{
		"a":true,
		"b":true,
		"c":true,
		"d":true,
		"e":true,
		"f":true,
		"g":true,
		"h":true,
	}

	if len(move) == 4 && mapBoard[string(move[0])] && mapBoard[string(move[2])] && string(move[1]) != "x"{
		return true
	}
	return false
}