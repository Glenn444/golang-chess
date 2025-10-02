import React from 'react'
import backgroundImage from '@/assets/chess-background.jpg';
import ChessBoard from './ChessBoard';


function ChessGame() {
  return (
    <div>
     <h1 className='text-xl'>Chess Game</h1>
      <div
       style={{
        backgroundImage: `linear-gradient(rgba(0,0,0,0.7), rgba(0,0,0,0.7)), url(${backgroundImage})`,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
        backgroundAttachment: 'fixed'
      }}
      className='w-[620px] h-[620px] mt-2'
    >
     
        <ChessBoard />
    </div>
    
    </div>
   
  )
}

export default ChessGame