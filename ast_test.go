package gqlparser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryDocMethods(t *testing.T) {
	doc, err := ParseQuery(`
		query Bob { foo { ...Frag } }
		fragment Frag on Foo {
			bar
		}
	`)

	require.Nil(t, err)
	t.Run("GetOperation", func(t *testing.T) {
		require.EqualValues(t, "Bob", doc.GetOperation("Bob").Name)
		require.Nil(t, doc.GetOperation("Alice"))
	})

	t.Run("GetFragment", func(t *testing.T) {
		require.EqualValues(t, "Frag", doc.GetFragment("Frag").Name)
		require.Nil(t, doc.GetOperation("Alice"))
	})
}
