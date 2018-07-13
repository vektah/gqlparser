package gqlparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestNamedTypeCompatability(t *testing.T) {
	assert.True(t, NamedType("A").IsCompatible(NamedType("A")))
	assert.False(t, NamedType("A").IsCompatible(NamedType("B")))

	assert.True(t, ListType{NamedType("A")}.IsCompatible(ListType{NamedType("A")}))
	assert.False(t, ListType{NamedType("A")}.IsCompatible(ListType{NamedType("B")}))
	assert.False(t, ListType{NamedType("A")}.IsCompatible(ListType{NamedType("B")}))

	assert.True(t, ListType{NamedType("A")}.IsCompatible(ListType{NamedType("A")}))
	assert.False(t, ListType{NamedType("A")}.IsCompatible(ListType{NamedType("B")}))
	assert.False(t, ListType{NamedType("A")}.IsCompatible(ListType{NamedType("B")}))

	assert.True(t, NonNullType{NamedType("A")}.IsCompatible(NamedType("A")))
	assert.False(t, NamedType("A").IsCompatible(NonNullType{NamedType("A")}))

	assert.True(t, NonNullType{ListType{NamedType("String")}}.IsCompatible(NonNullType{ListType{NamedType("String")}}))
	assert.True(t, NonNullType{ListType{NamedType("String")}}.IsCompatible(ListType{NamedType("String")}))
	assert.False(t, ListType{NamedType("String")}.IsCompatible(NonNullType{ListType{NamedType("String")}}))

	assert.True(t, ListType{NonNullType{NamedType("String")}}.IsCompatible(ListType{NamedType("String")}))
	assert.False(t, ListType{NamedType("String")}.IsCompatible(ListType{NonNullType{NamedType("String")}}))
}
