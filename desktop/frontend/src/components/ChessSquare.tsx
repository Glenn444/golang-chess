import React from 'react'
import alpha from '../assets/piece/alpha/index'
import { GameState } from '../App';

function ChessSquare({ file, rank, game }: { file: string, rank: number, game: GameState }) {
    const fileNumber = file.charCodeAt(0) - 'a'.charCodeAt(0) + 1;
    const isLight = (fileNumber + rank) % 2 == 0;
    const background = isLight ? '' : 'bg-amber-800';

    // Convert file/rank to board indices
    const col = fileNumber - 1; // 'a'=0, 'b'=1, etc.
    const row = 8 - rank;        // rank 8=0, rank 1=7 (assuming standard chess board indexing)

    // Get the square at this position
    const square = game.Board[row]?.[col];
    
    // Get the piece image if square is occupied
    let pieceImage = null;
    if (square?.Occupied && square.Piece) {
        // Construct the alpha key: first char = color, second char = piece type
        // Example: 'bB' for black bishop, 'wK' for white king
        const colorChar = square.Piece.Color; // 'b' or 'w'
        const pieceChar = square.Piece.PieceType[0].toUpperCase(); // 'B', 'K', 'N', etc.
        const alphaKey = `${colorChar}${pieceChar}`;
        pieceImage = alpha[alphaKey as keyof typeof alpha];
    }

    const handleClick = (game:GameState)=>{
        console.log(game.Board[row][col].Occupied);
        
        console.log('clicked');
        
    }
    return (
        <div className={`w-[77.5px] h-[77.5px] ${background} flex justify-center items-center`} onClick={()=>handleClick(game)}>
            {pieceImage ? (
                <img src={pieceImage} alt="chess piece" className="w-full h-full" />
            ) : (
                <span>{file}-{rank}</span>
            )}
        </div>
    )
}

export default ChessSquare