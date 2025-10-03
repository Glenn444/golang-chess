import { useEffect, useState } from 'react';
import logo from './assets/images/logo-universal.png';
import { GetBoardState, GetCurrentPlayer, MakeMove } from "../wailsjs/go/main/App";


import ChessGame from './components/ChessGame';


interface Square{
    Occupied: boolean;
	Piece:  Piece
}
export interface GameState{
    Board: Square[][];
    CurrentPlayer: "w" | "b";
}



interface Piece {
  getLegalSquares(): string[];
  getColor(): string;
  getPosition(): string;
  getPieceType(): string;
  assignPosition(pos: string): void;
  toString(): string;
}

function App() {
    const [gamestate, setGameState] = useState<GameState[]>([])

    const updateGameState = (game: GameState[]) => setGameState(game)

    useEffect(() => {

        GetBoardState()
            .then(updateGameState)
            .catch(error => {
                console.error("Error fetching board state:", error);
            });
    }, []);

    function greet() {

        console.log(gamestate);
    }


    return (
        <div id="App" className='flex justify-center'>
            <button onClick={greet}>Game</button>
            <ChessGame game={gamestate} />
        </div>
    )
}

export default App
