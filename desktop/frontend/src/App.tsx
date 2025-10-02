import {useState} from 'react';
import logo from './assets/images/logo-universal.png';


import ChessGame from './components/ChessGame';


function App() {
    const [resultText, setResultText] = useState(0);
    const [num, setNumber] = useState(0);
    const updateName = (e: any) => setNumber(Number(e.target.value));
    const updateResultText = (result: number) => setResultText(result);

    return (
        <div id="App" className='flex justify-center'>
            <ChessGame />
        </div>
    )
}

export default App
