package api

import (
	"context"
	"time"

	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/jackc/pgx/v5/pgtype"
)

// ── Shared response types ──────────────────────────────────────────

type GameResponse struct {
	ID              string    `json:"id"`
	WhitePlayerID   string    `json:"white_player_id"`
	BlackPlayerID   string    `json:"black_player_id"`
	WhitePlayerName string    `json:"white_player_name"`
	BlackPlayerName string    `json:"black_player_name"`
	State           string    `json:"state"`
	InCheck         bool      `json:"in_check"`
	CurrentPlayer   string    `json:"current_player"`
	MoveCount       int32     `json:"move_count"`
	BoardState      string    `json:"board_state"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type GameMoveResponse struct {
	ID           string    `json:"id"`
	GameID       string    `json:"game_id"`
	PlayerID     string    `json:"player_id"`
	PlayerColor  string    `json:"player_color"`
	MoveNotation string    `json:"move_notation"`
	MoveNumber   int32     `json:"move_number"`
	CreatedAt    time.Time `json:"created_at"`
}

type GameStatePayload struct {
	Game             *pieces.GameState `json:"game"`
	OpponentUsername string            `json:"opponent_username"`
}

// ── Mappers ────────────────────────────────────────────────────────

func (server *Server) toGameResponse(ctx context.Context, g db.Game) GameResponse {
	return GameResponse{
		ID:              uidStr(g.ID),
		WhitePlayerID:   uidStr(g.WhitePlayerID),
		BlackPlayerID:   uidStr(g.BlackPlayerID),
		WhitePlayerName: server.lookupUsername(ctx, g.WhitePlayerID),
		BlackPlayerName: server.lookupUsername(ctx, g.BlackPlayerID),
		State:           string(g.State),
		InCheck:         g.InCheck,
		CurrentPlayer:   string(g.CurrentPlayer),
		MoveCount:       g.MoveCount,
		BoardState:      g.BoardState,
		CreatedAt:       g.CreatedAt.Time,
		UpdatedAt:       g.UpdatedAt.Time,
	}
}

func (server *Server) toGameResponses(ctx context.Context, games []db.Game) []GameResponse {
	out := make([]GameResponse, len(games))
	for i, g := range games {
		out[i] = server.toGameResponse(ctx, g)
	}
	return out
}

// lookupUsername returns the username for a user ID, or empty string if
// the ID is invalid or the user can't be found.
func (server *Server) lookupUsername(ctx context.Context, id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	user, err := server.store.GetUserByID(ctx, id)
	if err != nil {
		return ""
	}
	return user.Username
}

// uidStr converts a pgtype.UUID to its string representation.
func uidStr(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	u, err := id.UUIDValue()
	if err != nil {
		return ""
	}
	return u.String()
}

func toGameMoveResponses(moves []db.GameMove) []GameMoveResponse {
	out := make([]GameMoveResponse, len(moves))
	for i, m := range moves {
		out[i] = GameMoveResponse{
			ID:           m.ID.String(),
			GameID:       m.GameID.String(),
			PlayerID:     m.PlayerID.String(),
			PlayerColor:  string(m.PlayerColor),
			MoveNotation: m.MoveNotation,
			MoveNumber:   m.MoveNumber,
			CreatedAt:    m.CreatedAt.Time,
		}
	}
	return out
}
