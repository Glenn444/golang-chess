import React from 'react'
import backgroundImage from '@/assets/chess-background.jpg';
import ChessBoard from './ChessBoard';
import { GameState } from '../App';


function ChessGame({game}:{game:GameState[]}) {
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
     
        <ChessBoard game={game}/>
    </div>
    
    </div>
   
  )
}

export default ChessGame