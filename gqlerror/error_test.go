package gqlerror

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorFormatting(t *testing.T) {
	t.Run("without filename", func(t *testing.T) {
		err := ErrorLocf("", 66, 2, "kabloom")

		require.Equal(t, `input:66: kabloom`, err.Error())
		require.Equal(t, nil, err.Extensions["file"])
	})

	t.Run("with filename", func(t *testing.T) {
		err := ErrorLocf("schema.graphql", 66, 2, "kabloom")

		require.Equal(t, `schema.graphql:66: kabloom`, err.Error())
		require.Equal(t, "schema.graphql", err.Extensions["file"])
	})

	t.Run("with path", func(t *testing.T) {
		err := ErrorPathf([]interface{}{"a", 1, "b"}, "kabloom")

		require.Equal(t, `input: a[1].b kabloom`, err.Error())
	})
}
