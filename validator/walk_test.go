package validator

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

func TestWalker(t *testing.T) {
	schema, err := LoadSchema(Prelude, &ast.Source{Input: "type Query { name: String }\n schema { query: Query }"})
	require.NoError(t, err)
	query, err := parser.ParseQuery(&ast.Source{Input: "{ as: name }"})
	require.NoError(t, err)

	called := false
	observers := &Events{}
	observers.OnField(func(walker *Walker, field *ast.Field) {
		called = true

		require.Equal(t, "name", field.Name)
		require.Equal(t, "as", field.Alias)
		require.Equal(t, "name", field.Definition.Name)
		require.Equal(t, "Query", field.ObjectDefinition.Name)
	})

	Walk(schema, query, observers)

	require.True(t, called)
}

func TestWalkInlineFragment(t *testing.T) {
	schema, err := LoadSchema(Prelude, &ast.Source{Input: "type Query { name: String }\n schema { query: Query }"})
	require.NoError(t, err)
	query, err := parser.ParseQuery(&ast.Source{Input: "{ ... { name } }"})
	require.NoError(t, err)

	called := false
	observers := &Events{}
	observers.OnField(func(walker *Walker, field *ast.Field) {
		called = true

		require.Equal(t, "name", field.Name)
		require.Equal(t, "name", field.Definition.Name)
		require.Equal(t, "Query", field.ObjectDefinition.Name)
	})

	Walk(schema, query, observers)

	require.True(t, called)
}
