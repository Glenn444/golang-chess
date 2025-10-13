
function IndicesToChessNotation(row:number,col:number):string{
    let boardLetter = ["a","b","c","d","e","f","g","h"]
    let letter = boardLetter[col]

    const pos = `${letter}${row+1}`
    return pos
}

export default IndicesToChessNotation
