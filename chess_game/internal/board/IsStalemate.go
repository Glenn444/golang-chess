package board

import "github.com/Glenn444/golang-chess/internal/pieces"

// IsStalemate returns true when the current player is not in check but has
// no legal move.
func IsStalemate(game pieces.GameState) bool {
	if IsKinginCheck(game) {
		return false
	}
	return !currentPlayerHasLegalMove(game)
}
