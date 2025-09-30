import React from 'react'
import backgroundImage from '@/assets/chess-background.jpg';


function ChessBoard() {
  return (
    <div
       style={{
        backgroundImage: `linear-gradient(rgba(0,0,0,0.7), rgba(0,0,0,0.7)), url(${backgroundImage})`,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
        backgroundAttachment: 'fixed'
      }}
      className='w-[620px] h-[620px]'
    >
        ChessBoard
    </div>
  )
}

export default ChessBoard