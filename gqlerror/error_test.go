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

	require.Equal(t, `input:66 kabloom`, err.Error())
}
