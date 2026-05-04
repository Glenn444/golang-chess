package board

import (
	"testing"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/internal/utils/chess"
)


func TestMove(t *testing.T) {
	moveTests := []struct {
		name          string
		boardState    map[string]string
		currentPlayer string
		castling      pieces.Castling
		move          string
		expectedErr   bool
		afterMove     func(t *testing.T, game *pieces.GameState) // optional post-move assertions
	}{
		// ─── Invalid / edge cases ───────────────────────────────────────────
		{
			name:          "Move too short — invalid",
			boardState:    map[string]string{"e2": "P"},
			currentPlayer: "w",
			move:          "e",
			expectedErr:   true,
		},
		{
			name:          "Empty move string",
			boardState:    map[string]string{"e2": "P"},
			currentPlayer: "w",
			move:          "",
			expectedErr:   true,
		},

		// ─── Castling delegation ────────────────────────────────────────────
		{
			name:          "Delegates to CastlingMove — white kingside",
			boardState:    map[string]string{"e1": "K", "h1": "R"},
			currentPlayer: "w",
			castling:      pieces.Castling{},
			move:          "e1g1",
			expectedErr:   false,
		},
		{
			name:          "Delegates to CastlingMove — fails when king moved",
			boardState:    map[string]string{"e1": "K", "h1": "R"},
			currentPlayer: "w",
			castling:      pieces.Castling{WhiteKingMoved: true},
			move:          "e1g1",
			expectedErr:   true,
		},

		// ─── Pawn moves ─────────────────────────────────────────────────────
		{
			name:          "Pawn move e2 to e4",
			boardState:    map[string]string{"e2": "P", "e1": "K", "e8": "k"},
			currentPlayer: "w",
			move:          "e4",
			expectedErr:   false,
			afterMove: func(t *testing.T, game *pieces.GameState) {
				row, col, _ := chess.ChessNotationToIndices("e4")
				if !game.Board[row][col].Occupied {
					t.Error("expected e4 to be occupied after pawn move")
				}
				row, col, _ = chess.ChessNotationToIndices("e2")
				if game.Board[row][col].Occupied {
					t.Error("expected e2 to be empty after pawn move")
				}
			},
		},
		{
			name:          "Pawn move — no pawn to move",
			boardState:    map[string]string{"e1": "K", "e8": "k"},
			currentPlayer: "w",
			move:          "e4",
			expectedErr:   true,
		},

		// ─── Piece moves ────────────────────────────────────────────────────
		{
			name:          "Rook moves Ra1 to Ra4",
			boardState:    map[string]string{"a1": "R", "e1": "K", "e8": "k"},
			currentPlayer: "w",
			move:          "Ra4",
			expectedErr:   false,
			afterMove: func(t *testing.T, game *pieces.GameState) {
				row, col, _ := chess.ChessNotationToIndices("a4")
				if !game.Board[row][col].Occupied {
					t.Error("expected a4 to be occupied after rook move")
				}
				row, col, _ = chess.ChessNotationToIndices("a1")
				if game.Board[row][col].Occupied {
					t.Error("expected a1 to be empty after rook move")
				}
			},
		},
		{
			name:          "Knight moves Ng1 to Nf3",
			boardState:    map[string]string{"g1": "N", "e1": "K", "e8": "k"},
			currentPlayer: "w",
			move:          "Nf3",
			expectedErr:   false,
		},
		{
			name:          "Bishop moves Bc1 to Bf4",
			boardState:    map[string]string{"c1": "B", "e1": "K", "e8": "k"},
			currentPlayer: "w",
			move:          "Bf4",
			expectedErr:   false,
		},

		// ─── Capture moves ──────────────────────────────────────────────────
		{
			name:          "Pawn captures — exd5",
			boardState:    map[string]string{"e4": "P", "d5": "p", "e1": "K", "e8": "k"},
			currentPlayer: "w",
			move:          "exd5",
			expectedErr:   false,
			afterMove: func(t *testing.T, game *pieces.GameState) {
				row, col, _ := chess.ChessNotationToIndices("d5")
				if !game.Board[row][col].Occupied {
					t.Error("expected d5 to be occupied after capture")
				}
				row, col, _ = chess.ChessNotationToIndices("e4")
				if game.Board[row][col].Occupied {
					t.Error("expected e4 to be empty after capture")
				}
			},
		},
		{
			name:          "Rook captures — Rxd5",
			boardState:    map[string]string{"d1": "R", "d5": "p", "e1": "K", "e8": "k"},
			currentPlayer: "w",
			move:          "Rxd5",
			expectedErr:   false,
		},
		{
			name:          "Capture leaves king in check — rejected",
			boardState:    map[string]string{"e1": "K", "e4": "R", "e8": "r", "d4": "p"},
			currentPlayer: "w",
			move:          "Rxd4",
			expectedErr:   true,
		},

		// ─── Check constraints ──────────────────────────────────────────────
		{
			name:          "Move leaves king in check — rejected",
			boardState:    map[string]string{"e1": "K", "e4": "R", "e8": "r"},
			currentPlayer: "w",
			move:          "Ra4",
			expectedErr:   true,
		},

		// ─── Player switching ────────────────────────────────────────────────
		{
			name:          "Current player switches to black after white moves",
			boardState:    map[string]string{"e2": "P", "e1": "K", "e8": "k"},
			currentPlayer: "w",
			move:          "e4",
			expectedErr:   false,
			afterMove: func(t *testing.T, game *pieces.GameState) {
				if game.CurrentPlayer != "b" {
					t.Errorf("expected current player to be 'b', got '%s'", game.CurrentPlayer)
				}
			},
		},
		{
			name:          "Current player switches to white after black moves",
			boardState:    map[string]string{"e7": "p", "e1": "K", "e8": "k"},
			currentPlayer: "b",
			move:          "e5",
			expectedErr:   false,
			afterMove: func(t *testing.T, game *pieces.GameState) {
				if game.CurrentPlayer != "w" {
					t.Errorf("expected current player to be 'w', got '%s'", game.CurrentPlayer)
				}
			},
		},

		// ─── StockfishGame history ───────────────────────────────────────────
		{
			name:          "Move is appended to StockfishGame history",
			boardState:    map[string]string{"e2": "P", "e1": "K", "e8": "k"},
			currentPlayer: "w",
			move:          "e4",
			expectedErr:   false,
			afterMove: func(t *testing.T, game *pieces.GameState) {
				if len(game.StockfishGame) == 0 {
					t.Error("expected StockfishGame to have a move recorded")
				}
			},
		},
	}

	for _, tt := range moveTests {
		t.Run(tt.name, func(t *testing.T) {
			boardState := pieces.SetUpBoard(tt.boardState)
			gameState := &pieces.GameState{
				CurrentPlayer:  tt.currentPlayer,
				Board:          boardState,
				Castle:         tt.castling,
				CapturedPieces: make(map[string][]pieces.PieceInterface),
			}

			err := Move(gameState, tt.move)

			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error=%v, got error=%v (err: %v)", tt.expectedErr, err != nil, err)
			}

			if err == nil && tt.afterMove != nil {
				tt.afterMove(t, gameState)
			}
		})
	}
}