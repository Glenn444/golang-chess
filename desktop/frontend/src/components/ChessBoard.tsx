import React, { useState } from 'react'
import ChessSquare from './ChessSquare'
import { GameState } from '../App'

function ChessBoard({ game }: { game: GameState }) {
    const [activesquares, setActiveSquares] = useState<string[]>([]);

    let file = ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h']
    return (
        <div className='grid grid-cols-8 text-white'>
            {[8, 7, 6, 5, 4, 3, 2, 1].map((rank) => {
                return file.map((l, i) => {
                    return (
                        <ChessSquare 
                        setActiveSquares= {setActiveSquares}
                        activesquares = {activesquares}
                        game={game} 
                        file={l} 
                        rank={rank} 
                        key={`${l}${rank}`} 
                        />
                    )
                })
            })
            }
        </div>
    )
}

export default ChessBoard