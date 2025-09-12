package queen

import (
	"github.com/Glenn444/golang-chess/internal/bishop"
	"github.com/Glenn444/golang-chess/internal/rook"
)


func Get_legal_squares(pos string) []string {
	var positions []string
	rook_pos := rook.Get_legal_squares(pos)
	bishop_pos := bishop.Get_legal_squares(pos)
	
	positions = append(positions, rook_pos...)
	positions = append(positions, bishop_pos...)

	return positions
}