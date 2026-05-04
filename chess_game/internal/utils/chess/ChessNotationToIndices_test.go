package chess

import (
	"testing"

	"github.com/stretchr/testify/require"
)


func TestChessNotationtoIndices(t *testing.T){
	testCases := []struct{
		name string
		position string
		want struct{row,col int}
		wantErr bool
	}{
		{
			name: "Valid Position",
			position: "a1",
			want: struct{row,col int}{0,0},
			wantErr: false,
		},
		{
			name: "Invalid Position",
			position: "a",
			wantErr: true,
		},
	}

	for _, tc := range testCases{
		t.Run(tc.name,func(t *testing.T) {
			row,col,err := ChessNotationToIndices(tc.position)
			if tc.wantErr{
				require.Error(t,err)
				return
			}
			require.NoError(t,err)
			require.Equal(t,tc.want.row,row)
			require.Equal(t,tc.want.col,col)
		})
	}
}