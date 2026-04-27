package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token is expired")

	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Payload contains the payload data of the token
type Payload struct {
	jwt.RegisteredClaims
	ID        uuid.UUID
	Username  string    `json:"username"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
	TokenTpe  TokenType `json:"token_type"`
}

// NewPayload creates a new token payload with a specific username and duration
func NewPayload(username string,tokenType TokenType, duration time.Duration) (*Payload, error) {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	payload := &Payload{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        newUUID.String(),
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
		Username: username,
		ID:       newUUID,
		TokenTpe: TokenType(tokenType),
	}

	return payload, nil
}
