package board

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/internal/utils/chess"
)

var pieceLetters = []string{"B", "N", "Q", "R", "K"}
var promotionPieces = []string{"Q", "R", "B", "N"}

// Move applies a move for the current player. Accepted formats:
//   - coordinate (UCI): "e2e4", promotion "e7e8q"
//   - algebraic: "e4", "exd5", "Nf3", "Nxf3", "Nbd2", "N1d2", "e8=Q", "O-O", "O-O-O"
func Move(game1 *pieces.GameState, move string) error {
	move = strings.TrimSpace(move)
	if move == "O-O" || move == "O-O-O" {
		return CastlingMove(game1, move)
	}
	from, to, promo, err := resolveMove(game1, move)
	if err != nil {
		return err
	}
	return executeMove(game1, from, to, promo)
}

// resolveMove turns any accepted move format into an explicit source square,
// destination square and optional promotion piece.
func resolveMove(g *pieces.GameState, move string) (from, to, promo string, err error) {
	if len(move) < 2 {
		return "", "", "", errors.New("invalid move")
	}

	// Coordinate form carries the source square explicitly.
	if len(move) >= 4 && isSquare(move[0:2]) && isSquare(move[2:4]) {
		switch len(move) {
		case 4:
			return move[0:2], move[2:4], "", nil
		case 5:
			promo = strings.ToUpper(string(move[4]))
			if !slices.Contains(promotionPieces, promo) {
				return "", "", "", errors.New("invalid promotion piece")
			}
			return move[0:2], move[2:4], promo, nil
		}
		return "", "", "", errors.New("invalid move")
	}

	return resolveAlgebraic(g, move)
}

// resolveAlgebraic finds the unique source square for an algebraic move.
func resolveAlgebraic(g *pieces.GameState, move string) (string, string, string, error) {
	original := move

	promo := ""
	if i := strings.Index(move, "="); i != -1 {
		if i != len(move)-2 {
			return "", "", "", errors.New("invalid promotion notation")
		}
		promo = strings.ToUpper(string(move[len(move)-1]))
		if !slices.Contains(promotionPieces, promo) {
			return "", "", "", errors.New("invalid promotion piece")
		}
		move = move[:i]
	}
	if len(move) < 2 {
		return "", "", "", errors.New("invalid move")
	}

	pieceType := "P"
	body := move
	if slices.Contains(pieceLetters, string(move[0])) {
		pieceType = string(move[0])
		body = move[1:]
	}
	isCapture := strings.Contains(body, "x")
	body = strings.ReplaceAll(body, "x", "")

	disamb := ""
	var to string
	switch len(body) {
	case 2:
		to = body
	case 3:
		disamb = string(body[0])
		to = body[1:]
	default:
		return "", "", "", errors.New("invalid move")
	}
	if !isSquare(to) {
		return "", "", "", errors.New("invalid destination square")
	}
	_, toCol, err := chess.ChessNotationToIndices(to)
	if err != nil {
		return "", "", "", err
	}

	var candidates []string
	for _, row := range g.Board {
		for _, sq := range row {
			if !sq.Occupied || sq.Piece.GetColor() != g.CurrentPlayer || sq.Piece.GetPieceType() != pieceType {
				continue
			}
			pos := sq.Piece.GetPosition()
			// The disambiguator is either the source file letter or rank digit.
			if disamb != "" && !strings.Contains(pos, disamb) {
				continue
			}
			if !slices.Contains(sq.Piece.GetLegalSquares(g), to) {
				continue
			}
			// A pawn push stays on its file; a pawn capture leaves it.
			if pieceType == "P" {
				fromCol := int(pos[0] - 'a')
				if isCapture && fromCol == toCol {
					continue
				}
				if !isCapture && fromCol != toCol {
					continue
				}
			}
			candidates = append(candidates, pos)
		}
	}

	if len(candidates) == 0 {
		return "", "", "", fmt.Errorf("invalid move: %s", original)
	}
	if len(candidates) > 1 {
		return "", "", "", fmt.Errorf("ambiguous move %s: specify the source square (e.g. %s%s)", original, string(original[0]), candidates[0][0:1])
	}
	return candidates[0], to, promo, nil
}

// executeMove validates and applies a fully-resolved move. Nothing on game1
// is mutated unless the move is legal.
func executeMove(game1 *pieces.GameState, from, to, promo string) error {
	fromRow, fromCol, err := chess.ChessNotationToIndices(from)
	if err != nil {
		return err
	}
	toRow, toCol, err := chess.ChessNotationToIndices(to)
	if err != nil {
		return err
	}

	fromSq := game1.Board[fromRow][fromCol]
	if !fromSq.Occupied {
		return fmt.Errorf("no piece on %s", from)
	}
	piece := fromSq.Piece
	if piece.GetColor() != game1.CurrentPlayer {
		return fmt.Errorf("the piece on %s is not yours", from)
	}

	// A king moving two files from its home square is a castling attempt.
	if piece.GetPieceType() == "K" && (from == "e1" || from == "e8") && fromRow == toRow && abs(toCol-fromCol) == 2 {
		return CastlingMove(game1, from+to)
	}

	if !slices.Contains(piece.GetLegalSquares(game1), to) {
		return fmt.Errorf("invalid move: %s%s to %s", piece.GetPieceType(), from, to)
	}

	// Simulate the move on a board copy so an illegal move never mutates state.
	boardCopy := Create_board()
	CopyBoard(boardCopy, game1.Board)
	sim := &pieces.GameState{
		CurrentPlayer:   game1.CurrentPlayer,
		Board:           boardCopy,
		EnPassantTarget: game1.EnPassantTarget,
	}

	var captured pieces.PieceInterface
	capturedFrom := to
	if sim.Board[toRow][toCol].Occupied {
		captured = sim.Board[toRow][toCol].Piece
	}

	movingPiece := sim.Board[fromRow][fromCol].Piece
	isPawn := movingPiece.GetPieceType() == "P"

	// En passant: a pawn capturing diagonally onto the empty target square
	// removes the pawn that just double-pushed past it.
	if isPawn && captured == nil && fromCol != toCol && to == game1.EnPassantTarget {
		captured = sim.Board[fromRow][toCol].Piece
		capturedFrom = chess.Indices_to_chess_notation(fromRow, toCol)
		sim.Board[fromRow][toCol] = pieces.Square{}
	}

	movingPiece.AssignPosition(to)
	sim.Board[fromRow][fromCol] = pieces.Square{}
	sim.Board[toRow][toCol] = pieces.Square{Occupied: true, Piece: movingPiece}

	// Promotion: a pawn reaching the last rank becomes the chosen piece
	// (queen when none was specified).
	promoted := ""
	if isPawn && (toRow == 0 || toRow == 7) {
		if promo == "" {
			promo = "Q"
		}
		sim.Board[toRow][toCol] = pieces.Square{Occupied: true, Piece: newPromotedPiece(promo, movingPiece.GetColor(), to)}
		promoted = promo
	} else if promo != "" {
		return errors.New("promotion is only possible for a pawn reaching the last rank")
	}

	if IsKinginCheck(sim) {
		return errors.New("move leaves your king in check")
	}

	// The move is legal — commit it to the real game state.
	CopyBoard(game1.Board, sim.Board)
	if captured != nil {
		if game1.CapturedPieces == nil {
			game1.CapturedPieces = make(map[string][]pieces.PieceInterface)
		}
		game1.CapturedPieces[game1.CurrentPlayer] = append(game1.CapturedPieces[game1.CurrentPlayer], captured)
		revokeCastlingRightsOnCapture(game1, captured, capturedFrom)
	}
	pieces.CastlePieceMoved(game1, piece.GetPieceType()+from)

	// A double pawn push exposes the skipped square to en passant for one ply.
	game1.EnPassantTarget = ""
	if isPawn && abs(toRow-fromRow) == 2 {
		game1.EnPassantTarget = chess.Indices_to_chess_notation((fromRow+toRow)/2, fromCol)
	}

	uci := from + to
	if promoted != "" {
		uci += strings.ToLower(promoted)
	}
	game1.StockfishGame = append(game1.StockfishGame, uci)

	switchPlayer(game1)
	return nil
}

// revokeCastlingRightsOnCapture clears the castling right tied to a rook
// captured on its original square.
func revokeCastlingRightsOnCapture(g *pieces.GameState, captured pieces.PieceInterface, square string) {
	if captured.GetPieceType() != "R" {
		return
	}
	switch square {
	case "a1":
		g.Castle.WhiteRookQueensideMoved = true
	case "h1":
		g.Castle.WhiteRookKingsideMoved = true
	case "a8":
		g.Castle.BlackRookQueensideMoved = true
	case "h8":
		g.Castle.BlackRookKingsideMoved = true
	}
}

func newPromotedPiece(pieceType, color, pos string) pieces.PieceInterface {
	switch pieceType {
	case "R":
		return &pieces.Rook{PieceType: "R", Color: color, Position: pos, Points: 5}
	case "B":
		return &pieces.Bishop{PieceType: "B", Color: color, Position: pos, Points: 3}
	case "N":
		return &pieces.Knight{PieceType: "N", Color: color, Position: pos, Points: 3}
	default:
		return &pieces.Queen{PieceType: "Q", Color: color, Position: pos, Points: 9}
	}
}

func isSquare(s string) bool {
	return len(s) == 2 && s[0] >= 'a' && s[0] <= 'h' && s[1] >= '1' && s[1] <= '8'
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func switchPlayer(g *pieces.GameState) {
	if g.CurrentPlayer == "w" {
		g.CurrentPlayer = "b"
	} else {
		g.CurrentPlayer = "w"
	}
}
