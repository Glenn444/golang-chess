package board

import (
	"encoding/json"

	"github.com/Glenn444/golang-chess/internal/pieces"
)

// gameStateSnapshot stores everything needed to reconstruct a GameState from DB.
type gameStateSnapshot struct {
	BoardPositions map[string]string `json:"board"`
	StockfishGame  []string           `json:"stockfish_game"`
}

// SerializeGameState encodes the full board and move history as JSON.
func SerializeGameState(gs *pieces.GameState) string {
	snap := gameStateSnapshot{
		BoardPositions: snapshotBoard(gs.Board),
		StockfishGame:  gs.StockfishGame,
	}
	b, _ := json.Marshal(snap)
	return string(b)
}

// DeserializeGameState restores the board and move history from JSON.
// Other fields (CurrentPlayer, MoveNumber, etc.) come from the games table.
func DeserializeGameState(raw string) ([][]pieces.Square, []string) {
	if raw == "" {
		board := Initialise_board(Create_board())
		return board, nil
	}
	var snap gameStateSnapshot
	if err := json.Unmarshal([]byte(raw), &snap); err != nil {
		board := Initialise_board(Create_board())
		return board, nil
	}
	board := pieces.SetUpBoard(snap.BoardPositions)
	return board, snap.StockfishGame
}

// snapshotBoard converts a board to the map format used by SetUpBoard.
func snapshotBoard(board [][]pieces.Square) map[string]string {
	m := make(map[string]string)
	for _, row := range board {
		for _, sq := range row {
			if !sq.Occupied || sq.Piece == nil {
				continue
			}
			pieceChar := sq.Piece.GetPieceType()
			if sq.Piece.GetColor() == "b" {
				pieceChar = string(pieceChar[0] + 32) // uppercase → lowercase
			}
			m[sq.Piece.GetPosition()] = pieceChar
		}
	}
	return m
}

// BuildInitialBoardState returns the serialized starting position.
func BuildInitialBoardState() string {
	b := Initialise_board(Create_board())
	return serializeBoard(b)
}

// serializeBoard encodes just the board as JSON for initial board state.
func serializeBoard(board [][]pieces.Square) string {
	m := snapshotBoard(board)
	b, _ := json.Marshal(gameStateSnapshot{BoardPositions: m})
	return string(b)
}
