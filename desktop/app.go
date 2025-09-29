package main

import (
	"context"
	"fmt"


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
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}


// Greet returns a sum of any given number
func (a *App) Add(num int) int {
	return num * 2
}