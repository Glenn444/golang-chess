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
    CurrentPlayer: string;
}



export interface Piece {
  PieceType: string;
  Color: string;
  Position: string;
}

function App() {
    const emptyState: GameState = {
    Board: Array(8).fill(null).map(() => 
        Array(8).fill(null).map(() => ({
            Occupied: false,
            Piece: null as any
        }))
    ),
    CurrentPlayer: "w"
};
   
    const [gamestate, setGameState] = useState<GameState>(emptyState)

    const updateGameState = (game: GameState) => setGameState(game)

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
