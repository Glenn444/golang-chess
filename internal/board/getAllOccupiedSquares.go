package board

import "github.com/Glenn444/golang-chess/utils"

func GetAllOccupiedSquares(game GameState) []string {
	var OccupiedSquares []string
	for i, b := range game.Board {
		for j, s := range b {
			if s.Occupied && game.CurrentPlayer == s.Piece.GetColor() {
				pos := utils.Indices_to_chess_notation(i, j)
				OccupiedSquares = append(OccupiedSquares, pos)
			}
		}
	}

	return OccupiedSquares
}
