package api

import "time"

type ChatMessageResponse struct {
	ID        string    `json:"id"`
	GameID    string    `json:"game_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type VoiceSessionResponse struct {
	ID          string    `json:"id"`
	GameID      string    `json:"game_id"`
	InitiatorID string    `json:"initiator_id"`
	State       string    `json:"state"`
	StartedAt   time.Time `json:"started_at"`
	EndedAt     time.Time `json:"ended_at"`
}

type CheckUsernameResponse struct {
	Username string `json:"username"`
	Exists   bool   `json:"exists"`
}

type DeletedResponse struct {
	Message string `json:"message"`
}