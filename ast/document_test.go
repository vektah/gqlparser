package ast_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/parser"
)

func TestQueryDocMethods(t *testing.T) {
	doc, err := parser.ParseQuery(&Source{Input: `
		query Bob { foo { ...Frag } }
		fragment Frag on Foo {
			bar
		}
	`})

	require.Nil(t, err)
	t.Run("GetOperation", func(t *testing.T) {
		require.EqualValues(t, "Bob", doc.Operations.ForName("Bob").Name)
		require.Nil(t, doc.Operations.ForName("Alice"))
	})

	t.Run("GetFragment", func(t *testing.T) {
		require.EqualValues(t, "Frag", doc.Fragments.ForName("Frag").Name)
		require.Nil(t, doc.Fragments.ForName("Alice"))
	})
}

func TestNamedTypeCompatability(t *testing.T) {
	assert.True(t, NamedType("A").IsCompatible(NamedType("A")))
	assert.False(t, NamedType("A").IsCompatible(NamedType("B")))

	assert.True(t, ListType(NamedType("A")).IsCompatible(ListType(NamedType("A"))))
	assert.False(t, ListType(NamedType("A")).IsCompatible(ListType(NamedType("B"))))
	assert.False(t, ListType(NamedType("A")).IsCompatible(ListType(NamedType("B"))))

	assert.True(t, ListType(NamedType("A")).IsCompatible(ListType(NamedType("A"))))
	assert.False(t, ListType(NamedType("A")).IsCompatible(ListType(NamedType("B"))))
	assert.False(t, ListType(NamedType("A")).IsCompatible(ListType(NamedType("B"))))

	assert.True(t, NonNullNamedType("A").IsCompatible(NamedType("A")))
	assert.False(t, NamedType("A").IsCompatible(NonNullNamedType("A")))

	assert.True(t, NonNullListType(NamedType("String")).IsCompatible(NonNullListType(NamedType("String"))))
	assert.True(t, NonNullListType(NamedType("String")).IsCompatible(ListType(NamedType("String"))))
	assert.False(t, ListType(NamedType("String")).IsCompatible(NonNullListType(NamedType("String"))))

	assert.True(t, ListType(NonNullNamedType("String")).IsCompatible(ListType(NamedType("String"))))
	assert.False(t, ListType(NamedType("String")).IsCompatible(ListType(NonNullNamedType("String"))))
}
