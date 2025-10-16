import React, { useEffect, useState } from 'react'
import alpha from '../assets/piece/alpha/index'
import { GameState, Piece } from '../App';
import IndicesToChessNotation from '../utils/IndicesToChessNotation';
import { GetLegalSquares } from '../../wailsjs/go/main/App';

function ChessSquare({ file, rank, game, activesquares, setActiveSquares,setSelectedPiece,selectedPiece }:
    { file: string, rank: number, game: GameState, activesquares: string[], selectedPiece:Piece | undefined,
        setActiveSquares: React.Dispatch<React.SetStateAction<string[]>>,
        setSelectedPiece:React.Dispatch<React.SetStateAction<Piece | undefined>>
     }) {
    

    const fileNumber = file.charCodeAt(0) - 'a'.charCodeAt(0) + 1;
    const isLight = (fileNumber + rank) % 2 == 0;
    const background = isLight ? '' : 'bg-amber-800';


    // Convert file/rank to board indices
    const col = fileNumber - 1; // 'a'=0, 'b'=1, etc.
    const row = rank - 1;        // rank 8=0, rank 1=7 (assuming standard chess board indexing)

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

    const handleClick = async (game: GameState) => {
        console.log("clicked piece: ", game.Board[row][col].Piece == selectedPiece);
        if (game.Board[row][col].Occupied && game.Board[row][col].Piece == selectedPiece) {
            setActiveSquares([])
           setSelectedPiece(undefined)
        }else if (game.Board[row][col].Occupied) {
             
        const legalSquares = await GetLegalSquares(row, col);

        //console.log("Legal squares returned:", legalSquares);
        setActiveSquares(legalSquares);
            setSelectedPiece(game.Board[row][col].Piece)
        }

       
    }



    const currentSquare = `${file}${rank}`;
    const isActiveSquare = activesquares.includes(currentSquare);
    //const isActiveSquare = () => activesquares.includes(`${file}${rank}`)
    const isActiveBackground = isActiveSquare ? 'bg-gray-50/40' : ''

    return (
        <div className={`w-[77.5px] h-[77.5px] ${background} ${isActiveBackground} flex justify-center items-center`} onClick={() => handleClick(game)}>
            {pieceImage ? (
                <img src={pieceImage} alt="chess piece" className="w-full h-full" />
            ) : (
                <span>{file}{rank}</span>
            )}
        </div>
    )
}

export default ChessSquare