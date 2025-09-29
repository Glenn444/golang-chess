# Golang Chess Game

A desktop chess game built with [Wails](https://wails.io/) for the UI and Go for the CLI and backend logic.

## Features

- Play chess against the computer or another player
- Desktop GUI powered by Wails
- Command-line interface for quick games
- Save and load games
- Move validation and game status updates

## Requirements

- Go 1.20+
- Node.js (for Wails)
- Wails CLI (`npm install -g wails`)

## Getting Started

### Desktop App

```bash
# Install dependencies
wails dev
# Build the desktop app
wails build
# Run the app
./build/bin/golang-chess
```

### CLI

```bash
go run main.go
```

## Project Structure

- `/frontend` - Wails frontend (UI)
- `/backend` - Go backend (game logic)
- `main.go` - CLI entry point

## License

MIT
