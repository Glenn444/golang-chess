package bishop


func Get_legal_squares(pos string) []string  {
	var positions []string
	pos_top_left := get_horizontal_squares_top_left(pos)
	pos_top_right := get_horizontal_squares_top_right(pos)

	positions = append(positions, pos_top_left...)
	positions = append(positions, pos_top_right...)

	return  positions
}