package fieldmap

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/ast"
)

func TestFieldMap(t *testing.T) {
	var m Map

	t.Run("zero wont panic", func(t *testing.T) {
		f, ok := m.Get("bob")
		require.False(t, ok)
		require.Nil(t, f)
		require.Equal(t, 0, m.Len())
	})

	m.Push("a.a", &Node{&ast.Field{Name: "0"}, []interface{}{1, "2"}})
	m.Push("a.b", &Node{&ast.Field{Name: "1"}, []interface{}{2, "1"}})
	m.Push("a.b", &Node{&ast.Field{Name: "2"}, []interface{}{3, "1"}})

	t.Run("get", func(t *testing.T) {
		t.Run("fetches values", func(t *testing.T) {
			f, ok := m.Get("a.a")
			require.True(t, ok)
			require.Equal(t, "0", f[0].Field.Name)
		})

		t.Run("collects duplicates in order", func(t *testing.T) {
			f, ok := m.Get("a.b")
			require.True(t, ok)
			require.Equal(t, "1", f[0].Field.Name)
			require.Equal(t, "2", f[1].Field.Name)
		})

		t.Run("returns not found", func(t *testing.T) {
			f, ok := m.Get("bob")
			require.False(t, ok)
			require.Nil(t, f)
		})
	})

	t.Run("at", func(t *testing.T) {
		t.Run("fetches values", func(t *testing.T) {
			key, value := m.At(0)
			require.Equal(t, "a.a", key)
			require.Equal(t, "0", value[0].Field.Name)
		})

		t.Run("collects duplicates in order", func(t *testing.T) {
			key, value := m.At(1)
			require.Equal(t, "a.b", key)
			require.Equal(t, "1", value[0].Field.Name)
			require.Equal(t, "2", value[1].Field.Name)
		})

		t.Run("panics on out of bounds", func(t *testing.T) {
			require.Panics(t, func() {
				m.At(100)
			})
		})
	})

	t.Run("len", func(t *testing.T) {
		require.Equal(t, 2, m.Len())
	})

	t.Run("reset", func(t *testing.T) {
		m.Reset()
		f, ok := m.Get("bob")
		require.False(t, ok)
		require.Nil(t, f)
		require.Equal(t, 0, m.Len())
	})
}
