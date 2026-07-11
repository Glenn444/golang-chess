package board

import (
	"testing"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/internal/utils/chess"
)

func newGame(boardState map[string]string, currentPlayer string) *pieces.GameState {
	return &pieces.GameState{
		CurrentPlayer:  currentPlayer,
		Board:          pieces.SetUpBoard(boardState),
		CapturedPieces: make(map[string][]pieces.PieceInterface),
	}
}

func pieceAt(t *testing.T, g *pieces.GameState, pos string) pieces.PieceInterface {
	t.Helper()
	row, col, err := chess.ChessNotationToIndices(pos)
	if err != nil {
		t.Fatalf("bad square %s: %v", pos, err)
	}
	if !g.Board[row][col].Occupied {
		t.Fatalf("expected a piece on %s", pos)
	}
	return g.Board[row][col].Piece
}

func assertEmpty(t *testing.T, g *pieces.GameState, pos string) {
	t.Helper()
	row, col, _ := chess.ChessNotationToIndices(pos)
	if g.Board[row][col].Occupied {
		t.Errorf("expected %s to be empty, found %s", pos, g.Board[row][col].Piece.GetPieceType())
	}
}

// Pawns must move by color, not by rank: a black pawn on rank 2 moves toward
// rank 1 (and promotes), never back up the board.
func TestBlackPawnOnRank2MovesDownAndPromotes(t *testing.T) {
	g := newGame(map[string]string{"b2": "p", "e8": "k", "e1": "K"}, "b")

	if err := Move(g, "b2b3"); err == nil {
		t.Error("black pawn must not move backwards from b2 to b3")
	}
	if err := Move(g, "b2b1"); err != nil {
		t.Fatalf("black pawn should promote on b1: %v", err)
	}
	p := pieceAt(t, g, "b1")
	if p.GetPieceType() != "Q" || p.GetColor() != "b" {
		t.Errorf("expected a black queen on b1, got %s %s", p.GetColor(), p.GetPieceType())
	}
}

func TestWhitePromotionChoices(t *testing.T) {
	g := newGame(map[string]string{"e7": "P", "a8": "k", "e1": "K"}, "w")
	if err := Move(g, "e7e8n"); err != nil {
		t.Fatalf("underpromotion failed: %v", err)
	}
	if p := pieceAt(t, g, "e8"); p.GetPieceType() != "N" {
		t.Errorf("expected knight on e8, got %s", p.GetPieceType())
	}

	g = newGame(map[string]string{"e7": "P", "a8": "k", "e1": "K"}, "w")
	if err := Move(g, "e8=Q"); err != nil {
		t.Fatalf("algebraic promotion failed: %v", err)
	}
	if p := pieceAt(t, g, "e8"); p.GetPieceType() != "Q" {
		t.Errorf("expected queen on e8, got %s", p.GetPieceType())
	}
}

// A pawn capture must come from a diagonally adjacent square — it may not
// teleport across the board, and an impossible capture must error, not panic.
func TestPawnCaptureValidated(t *testing.T) {
	g := newGame(map[string]string{"e4": "P", "a5": "r", "e1": "K", "e8": "k"}, "w")
	if err := Move(g, "exa5"); err == nil {
		t.Error("pawn on e4 must not capture a5")
	}
	if err := Move(g, "bxa6"); err == nil {
		t.Error("capture with no capturing pawn must fail")
	}
	pieceAt(t, g, "e4") // pawn untouched
}

// A coordinate move must move the piece on its source square, not the first
// piece of that type that can reach the destination.
func TestCoordinateMoveUsesSourceSquare(t *testing.T) {
	g := newGame(map[string]string{"c3": "N", "g1": "N", "a1": "K", "a8": "k"}, "w")
	if err := Move(g, "c3e2"); err != nil {
		t.Fatalf("c3e2 failed: %v", err)
	}
	assertEmpty(t, g, "c3")
	pieceAt(t, g, "g1")
	pieceAt(t, g, "e2")
}

func TestAmbiguousAlgebraicMoveRejected(t *testing.T) {
	g := newGame(map[string]string{"c3": "N", "g1": "N", "a1": "K", "a8": "k"}, "w")
	if err := Move(g, "Ne2"); err == nil {
		t.Error("Ne2 with two candidate knights must be rejected as ambiguous")
	}
	if err := Move(g, "Nce2"); err != nil {
		t.Errorf("file-disambiguated Nce2 should work: %v", err)
	}
	assertEmpty(t, g, "c3")
}

func TestRankDisambiguation(t *testing.T) {
	g := newGame(map[string]string{"a1": "R", "a5": "R", "h2": "K", "h8": "k"}, "w")
	if err := Move(g, "R1a3"); err != nil {
		t.Fatalf("rank-disambiguated R1a3 should work: %v", err)
	}
	assertEmpty(t, g, "a1")
	pieceAt(t, g, "a5")
}

// e1g1 is only castling when the king stands on e1 — a rook sliding e1→g1 is
// an ordinary move.
func TestRookE1ToG1IsNotCastling(t *testing.T) {
	g := newGame(map[string]string{"e1": "R", "d1": "K", "d8": "k"}, "w")
	if err := Move(g, "e1g1"); err != nil {
		t.Fatalf("rook move e1g1 should be legal: %v", err)
	}
	if p := pieceAt(t, g, "g1"); p.GetPieceType() != "R" {
		t.Errorf("expected rook on g1, got %s", p.GetPieceType())
	}
}

// A rejected move must not revoke castling rights.
func TestRejectedMoveKeepsCastlingRights(t *testing.T) {
	g := newGame(map[string]string{"e1": "K", "h1": "R", "e8": "r", "a8": "k"}, "w")
	if err := Move(g, "e1e2"); err == nil {
		t.Fatal("Ke2 stays on the checked e-file and must be rejected")
	}
	if g.Castle.WhiteKingMoved {
		t.Error("rejected king move must not set WhiteKingMoved")
	}
}

// A king move that is a capture must revoke castling rights on the real game
// state (previously it was recorded on a discarded board copy).
func TestKingCaptureRevokesRights(t *testing.T) {
	g := newGame(map[string]string{"e1": "K", "d2": "p", "a8": "k"}, "w")
	if err := Move(g, "Kxd2"); err != nil {
		t.Fatalf("Kxd2 failed: %v", err)
	}
	if !g.Castle.WhiteKingMoved {
		t.Error("king capture must set WhiteKingMoved")
	}
}

// Castling must set the moved flags so a second castle is impossible, and
// queenside castling must land king on c-file and rook on d-file.
func TestCastlingSetsFlagsAndGeometry(t *testing.T) {
	g := newGame(map[string]string{"e1": "K", "a1": "R", "e8": "k", "a8": "r"}, "w")
	if err := Move(g, "O-O-O"); err != nil {
		t.Fatalf("white queenside castling failed: %v", err)
	}
	if p := pieceAt(t, g, "c1"); p.GetPieceType() != "K" {
		t.Errorf("expected king on c1, got %s", p.GetPieceType())
	}
	if p := pieceAt(t, g, "d1"); p.GetPieceType() != "R" {
		t.Errorf("expected rook on d1, got %s", p.GetPieceType())
	}
	assertEmpty(t, g, "a1")
	assertEmpty(t, g, "e1")
	if !g.Castle.WhiteKingMoved || !g.Castle.WhiteRookQueensideMoved {
		t.Error("castling must set the moved flags")
	}

	// Fresh position for black: after white's O-O-O above, the white rook on
	// d1 would legitimately attack d8 and forbid black's queenside castle.
	g = newGame(map[string]string{"e8": "k", "a8": "r", "e1": "K"}, "b")
	if err := Move(g, "O-O-O"); err != nil {
		t.Fatalf("black queenside castling failed: %v", err)
	}
	if p := pieceAt(t, g, "c8"); p.GetPieceType() != "K" {
		t.Errorf("expected king on c8, got %s", p.GetPieceType())
	}
	if p := pieceAt(t, g, "d8"); p.GetPieceType() != "R" {
		t.Errorf("expected rook on d8, got %s", p.GetPieceType())
	}
}

// A quiet-notation move onto an enemy piece is a capture and must be recorded.
func TestQuietMoveOntoEnemyRecordsCapture(t *testing.T) {
	g := newGame(map[string]string{"d1": "R", "d5": "p", "e1": "K", "e8": "k"}, "w")
	if err := Move(g, "Rd5"); err != nil {
		t.Fatalf("Rd5 failed: %v", err)
	}
	if len(g.CapturedPieces["w"]) != 1 {
		t.Errorf("expected 1 captured piece for white, got %d", len(g.CapturedPieces["w"]))
	}
}

// Capturing a rook on its original square removes that castling right.
func TestRookCapturedOnHomeSquareRevokesRights(t *testing.T) {
	g := newGame(map[string]string{"h1": "R", "h8": "r", "e1": "K", "a8": "k"}, "w")
	if err := Move(g, "Rxh8"); err != nil {
		t.Fatalf("Rxh8 failed: %v", err)
	}
	if !g.Castle.BlackRookKingsideMoved {
		t.Error("capturing the h8 rook must revoke black kingside castling")
	}
}

func TestEnPassant(t *testing.T) {
	g := newGame(map[string]string{"e5": "P", "d7": "p", "e1": "K", "e8": "k"}, "b")

	if err := Move(g, "d7d5"); err != nil {
		t.Fatalf("d7d5 failed: %v", err)
	}
	if g.EnPassantTarget != "d6" {
		t.Fatalf("expected en passant target d6, got %q", g.EnPassantTarget)
	}
	if err := Move(g, "exd6"); err != nil {
		t.Fatalf("en passant capture exd6 failed: %v", err)
	}
	if p := pieceAt(t, g, "d6"); p.GetPieceType() != "P" || p.GetColor() != "w" {
		t.Errorf("expected white pawn on d6, got %s %s", p.GetColor(), p.GetPieceType())
	}
	assertEmpty(t, g, "d5")
	assertEmpty(t, g, "e5")
	if len(g.CapturedPieces["w"]) != 1 {
		t.Errorf("expected the en passant victim to be recorded as captured")
	}
}

func TestEnPassantExpiresAfterOnePly(t *testing.T) {
	g := newGame(map[string]string{"e5": "P", "d7": "p", "h7": "p", "e1": "K", "e8": "k"}, "b")

	if err := Move(g, "d7d5"); err != nil {
		t.Fatalf("d7d5 failed: %v", err)
	}
	if err := Move(g, "e1d1"); err != nil { // white declines the capture
		t.Fatalf("Kd1 failed: %v", err)
	}
	if err := Move(g, "h7h6"); err != nil {
		t.Fatalf("h7h6 failed: %v", err)
	}
	if g.EnPassantTarget != "" {
		t.Errorf("en passant target should be cleared, got %q", g.EnPassantTarget)
	}
	if err := Move(g, "e5d6"); err == nil {
		t.Error("en passant must not be possible two plies later")
	}
}

// Castling rights and en-passant state must survive a serialize/deserialize
// round trip (previously every restored game regained full castling rights).
func TestSerializationRoundTripsCastlingAndEnPassant(t *testing.T) {
	g := newGame(map[string]string{"e1": "K", "h1": "R", "e8": "k"}, "w")
	g.Castle.WhiteKingMoved = true
	g.Castle.BlackRookQueensideMoved = true
	g.EnPassantTarget = "d6"
	g.StockfishGame = []string{"e2e4", "d7d5"}

	snap := DeserializeGameState(SerializeGameState(g))
	if !snap.Castle.WhiteKingMoved || !snap.Castle.BlackRookQueensideMoved {
		t.Error("castling rights lost in serialization round trip")
	}
	if snap.EnPassantTarget != "d6" {
		t.Errorf("en passant target lost, got %q", snap.EnPassantTarget)
	}
	if len(snap.StockfishGame) != 2 {
		t.Errorf("move history lost, got %v", snap.StockfishGame)
	}
}

// A full game played with coordinate moves (the format the WebSocket layer
// sends) must reach checkmate: the Scholar's Mate.
func TestScholarsMateViaCoordinateMoves(t *testing.T) {
	g := &pieces.GameState{
		CurrentPlayer:  "w",
		Board:          Initialise_board(Create_board()),
		CapturedPieces: make(map[string][]pieces.PieceInterface),
	}
	moves := []string{"e2e4", "e7e5", "f1c4", "b8c6", "d1h5", "g8f6", "h5f7"}
	for _, m := range moves {
		if err := Move(g, m); err != nil {
			t.Fatalf("move %s failed: %v", m, err)
		}
	}
	if !IsKinginCheck(g) {
		t.Error("black king should be in check after Qxf7")
	}
	if !IsCheckmate(g) {
		t.Error("Qxf7 should be checkmate")
	}
	if len(g.CapturedPieces["w"]) != 1 {
		t.Errorf("white should have captured exactly the f7 pawn, got %d captures", len(g.CapturedPieces["w"]))
	}
}

// Off-board notation like rank 0 or 9 must return an error, not indexes that
// panic downstream.
func TestNotationRejectsOffBoardRanks(t *testing.T) {
	for _, pos := range []string{"a0", "a9", "i1", "e10"} {
		if _, _, err := chess.ChessNotationToIndices(pos); err == nil {
			t.Errorf("expected %q to be rejected", pos)
		}
	}
}
