package errors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorFormatting(t *testing.T) {
	err := Syntax{
		Message: "kabloom",
		Locations: []Location{
			{Line: 66, Column: 33},
		},
	}

	require.Equal(t, `Syntax Error: kabloom (line 66, column 33)`, err.Error())
}
