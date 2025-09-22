package board

import (
	"fmt"

)

func Move(game *GameState, move string) {
	//pieces := []string{"Q","N","K","R","B"}
	//letter := string(move[0])
	pos := CurrentPlayer_Occupied_Piece_position(*game,move)
	fmt.Printf("Piece Position: %s\n", pos)

	
	// switch letter {
	// 	case "Q":
	// 		fmt.Printf("Letter: %s\n", letter)
	// 	case "N":
	// 		fmt.Printf("Knight: %s\n",letter)
	// 	case "K":
	// 		fmt.Printf("King: %s\n",letter)
	// 	case "R":
	// 		fmt.Printf("Rook: %s\n",letter)
	// 	case "B":
	// 		fmt.Printf("Bishop: %s\n",letter)
	// 	default:
	// 		fmt.Printf("Pawn: %s\n",letter)
	// }
	

}



// func Get_Piece()  {
	
// }