package pieces

import (
	"github.com/stretchr/testify/require"
	"testing"
)


func TestBishop(t *testing.T){
	b1 := Bishop{
		PieceType: "B",
		Color: "w",
		Position: "c1",
	}

	expectedlegalSquares := []string{"c1","b2","a3","d2","e3","f4","g5","h6"}
	gotlegalSquares := b1.GetLegalSquares()

	require.Equal(t,expectedlegalSquares,gotlegalSquares)
}