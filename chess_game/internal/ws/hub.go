package ws

import (
	"sync"

	"github.com/gofrs/uuid"
	"github.com/olahol/melody"
)

type Color int
type GameState int

const (
	w Color = iota
	b
)

const (
    Waiting    GameState = iota // Waiting for opponent to join
    Active                      // Game in progress
    Checkmate                   // Game over - king mated
    Stalemate                   // Game over - no legal moves, not in check
    Resign                      // Game over - player resigned
    Draw                        // Game over - agreed draw / repetition / 50-move rule
    Abandoned                   // Game over - player disconnected
)

//add a String() to make it "human-readable"
func (c Color) String()string{
	colors := [...]string{"w","b"}

	if c < 0 || int(c) >= len(colors){
		return "Unknown"
	}
	return colors[c]
}

func (gs GameState) String() string {
    states := [...]string{
        "Waiting",
        "Active",
        "Checkmate",
        "Stalemate",
        "Resign",
        "Draw",
        "Abandoned",
    }

    if gs < 0 || int(gs) >= len(states) {
        return "Unknown"
    }
    return states[gs]
}

type Player struct{
	UserId uuid.UUID
	Session *melody.Session
	GameId uuid.UUID
}
type Move struct{
	PlayerColor Color
	PieceMove string //e2e3
}

type Game struct{
	Id uuid.UUID
	Moves []Move
	Players [2]Player
	State GameState
	InCheck bool
}

var (
	games = map[uuid.UUID]*Game{}
	mu sync.RWMutex
)

