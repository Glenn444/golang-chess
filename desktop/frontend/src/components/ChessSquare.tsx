import React from 'react'
import alpha from '../assets/piece/alpha/index'
import { GameState } from '../App';


function ChessSquare({ file, rank,game }: { file: string, rank: number, game:GameState[] }) {
    const fileNumber = file.charCodeAt(0) - 'a'.charCodeAt(0) + 1;
    const isLight = (fileNumber + rank) % 2 == 0;
    const background = isLight ? '' : 'bg-amber-800';

    //alpha['bB'] black, Bishop
    return (
        <div className={`w-[77.5px] h-[77.5px] ${background}  flex justify-center items-center`}>
            {file}-{rank}
        </div>
    )
}

export default ChessSquare