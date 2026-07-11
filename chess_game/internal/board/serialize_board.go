package board

import (
	"encoding/json"

	"github.com/Glenn444/golang-chess/internal/pieces"
)

// gameStateSnapshot stores everything needed to reconstruct a GameState from DB.
type gameStateSnapshot struct {
	BoardPositions  map[string]string `json:"board"`
	StockfishGame   []string          `json:"stockfish_game"`
	Castle          pieces.Castling   `json:"castle"`
	EnPassantTarget string            `json:"en_passant_target,omitempty"`
}

// GameSnapshot is the decoded persisted engine state.
type GameSnapshot struct {
	Board           [][]pieces.Square
	StockfishGame   []string
	Castle          pieces.Castling
	EnPassantTarget string
}

// SerializeGameState encodes the board, move history, castling rights and
// en-passant state as JSON.
func SerializeGameState(gs *pieces.GameState) string {
	snap := gameStateSnapshot{
		BoardPositions:  snapshotBoard(gs.Board),
		StockfishGame:   gs.StockfishGame,
		Castle:          gs.Castle,
		EnPassantTarget: gs.EnPassantTarget,
	}
	b, _ := json.Marshal(snap)
	return string(b)
}

// DeserializeGameState restores the persisted engine state from JSON.
// Other fields (CurrentPlayer, MoveNumber, etc.) come from the games table.
func DeserializeGameState(raw string) GameSnapshot {
	var snap gameStateSnapshot
	if raw == "" || json.Unmarshal([]byte(raw), &snap) != nil {
		return GameSnapshot{Board: Initialise_board(Create_board())}
	}
	return GameSnapshot{
		Board:           pieces.SetUpBoard(snap.BoardPositions),
		StockfishGame:   snap.StockfishGame,
		Castle:          snap.Castle,
		EnPassantTarget: snap.EnPassantTarget,
	}
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
