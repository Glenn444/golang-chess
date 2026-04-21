package ws

import (
	"sync"

	"github.com/gofrs/uuid"
	"github.com/olahol/melody"
)

type Player struct{
	UserId uuid.UUID
	Session *melody.Session
	GameId uuid.UUID
}
type Move struct{
	PlayerColor string
	PieceMove string //Nf3

}

type Game struct{
	Id uuid.UUID
	Moves []Move
	Players [2]Player
	State string
}

var (
	games = map[uuid.UUID]*Game{}
	mu sync.RWMutex
)

