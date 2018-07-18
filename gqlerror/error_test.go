package gqlerror

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorFormatting(t *testing.T) {
	err := Error{
		Message: "kabloom",
		Locations: []Location{
			{Line: 66, Column: 33},
		},
	}

	require.Equal(t, `kabloom (line 66, column 33)`, err.Error())
}
