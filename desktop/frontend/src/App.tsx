import {useState} from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import {Greet,Add} from "../wailsjs/go/main/App";

function App() {
    const [resultText, setResultText] = useState(0);
    const [num, setNumber] = useState(0);
    const updateName = (e: any) => setNumber(Number(e.target.value));
    const updateResultText = (result: number) => setResultText(result);

    function greet() {
        Add(num).then(updateResultText);
    }

    return (
        <div id="App">
            <img src={logo} id="logo" alt="logo"/>
            <div id="result" className="result">{resultText}</div>
            <div id="input" className="input-box">
                <input id="name" className="input" onChange={updateName} autoComplete="off" name="input" type="text"/>
                <button className="btn" onClick={greet}>Sum</button>
            </div>
        </div>
    )
}

export default App
