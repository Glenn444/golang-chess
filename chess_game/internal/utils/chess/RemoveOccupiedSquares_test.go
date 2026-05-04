package chess

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveOccupiedSquares(t *testing.T){
	testCases := []struct{
		name string
		input []string
		remove []string
		want []string
	}{
		{
			name: "Valid",
			input: []string{"a","b","c","d"},
			want: []string{"a","b"},
			remove: []string{"c","d"},
		},
		{
			name: "Remove is Empty",
			input: []string{"a","b","c","d"},
			want: []string{"a","b","c","d"},
			remove: []string{},
		},
		{
			name: "Input is Empty",
			input: []string{},
			want: []string{},
			remove: []string{},
		},
		{
			name: "Input and Remove the same",
			input: []string{"a","b","c","d"},
			want: []string{},
			remove: []string{"a","b","c","d"},
		},
	}
	for _,tc := range testCases{
		
		t.Run(tc.name,func(t *testing.T) {
			squares := RemoveOwnOccupiedSquares(tc.input,tc.remove)
			
			require.Len(t,squares,len(tc.want))
			require.Equal(t,tc.want,squares)
			
		})
	}
}