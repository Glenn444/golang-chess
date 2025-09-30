package main

import (
	"context"


	"github.com/Glenn444/golang-chess/backend/pkg/board"
)

// App struct
type App struct {
	ctx context.Context
	game *board.GameState 
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	b := board.Create_board()
	initialBoard_position := board.Initialise_board(b)

	//Initialize your game
	a.game = &board.GameState{
		Board: initialBoard_position,
		CurrentPlayer: "w",
	}
}

func (a *App) MakeMove(move string) error {
	return board.Move(a.game,move)
}

func (a *App) GetBoardState() [][]board.Square{
	return a.game.Board
}