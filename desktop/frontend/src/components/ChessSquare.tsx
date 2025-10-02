import React from 'react'

function ChessSquare({ file, rank }: { file: string, rank: number }) {
    const fileNumber = file.charCodeAt(0) - 'a'.charCodeAt(0) + 1;
    const isLight = (fileNumber + rank) % 2 == 0;
    const background = isLight ? '' : 'bg-amber-800';

    
    return (
        <div className={`w-[77.5px] h-[77.5px] ${background}  flex justify-center items-center`}>
            {file}-{rank}
        </div>
    )
}

export default ChessSquare