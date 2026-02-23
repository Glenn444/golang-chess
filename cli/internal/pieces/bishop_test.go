package pieces

import (
	"testing"
)


func TestBishop(t *testing.T){
	b1 := Bishop{
		PieceType: "B",
		Color: "w",
		Position: "c1",
	}

	expectedlegalSquares := []string{"b2","a3","d2","e3","f4","g5","h6"}
	gotlegalSquares := b1.GetLegalSquares()

	equalSlices := compareSlices(expectedlegalSquares,gotlegalSquares)
	if equalSlices != true{
		t.Errorf("got %v want %v", gotlegalSquares, expectedlegalSquares)
	}
}

func compareSlices(expectedSlice[]string,gotSlice[]string)bool{
	finalSlice := []string{}
	for _,val1 := range gotSlice{
		for _, val2 := range expectedSlice{
			if val1 == val2{
				finalSlice = append(finalSlice,val1)
			}
		}
	}
	lengthExpectedSlice := len(expectedSlice)

	if lengthExpectedSlice == len(finalSlice){
		return true
	}
	return false
}