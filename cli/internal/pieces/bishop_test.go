package pieces

import (
	"testing"
)

func TestBishop(t *testing.T) {
	bishopTests := []struct {
		name         string
		piece        PieceInterface
		legalSquares []string
	}{
		{
			name: "bishop1",
			piece: &Bishop{
				PieceType: "B",
				Color:     "w",
				Position:  "c1",
			},
			legalSquares: []string{"b2", "a3", "d2", "e3", "f4", "g5", "h6"},
		},
		{
			name: "bishop2",
			piece: &Bishop{
				PieceType: "B",
				Color:     "w",
				Position:  "d4",
			},
			legalSquares: []string{"c3", "b2", "a1", "e5", "f6", "g7", "h8", "c5", "b6", "a7", "e3", "f2", "g1"},
		},
		{
			name: "bishop3",
			piece: &Bishop{
				PieceType: "B",
				Color:     "b",
				Position:  "a1",
			},
			legalSquares: []string{"b2", "c3", "d4", "e5", "f6", "g7", "h8"},
		},
		{
			name: "bishop3",
			piece: &Bishop{
				PieceType: "B",
				Color:     "b",
				Position:  "h8",
			},
			legalSquares: []string{"g7", "f6", "e5", "d4", "c3", "b2", "a1"},
		},
	}


	for _, tt := range bishopTests {
		t.Run(tt.name, func(t *testing.T) {
			gotlegalSquares := tt.piece.GetLegalSquares()
			equalSlices := compareSlices(tt.legalSquares,gotlegalSquares)
			if equalSlices != true{
				t.Errorf("got %v want %v",gotlegalSquares,tt.legalSquares)
			}
		})
	}
}

func compareSlices(expectedSlice []string, gotSlice []string) bool {
	finalSlice := []string{}
	for _, val1 := range gotSlice {
		for _, val2 := range expectedSlice {
			if val1 == val2 {
				finalSlice = append(finalSlice, val1)
			}
		}
	}
	lengthExpectedSlice := len(expectedSlice)

	if lengthExpectedSlice == len(finalSlice) {
		return true
	}
	return false
}
