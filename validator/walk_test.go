package validator

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/parser"
)

func TestWalker(t *testing.T) {
	schema, err := parser.LoadSchema("type Query { name: String }\n schema { query: Query }")
	require.Nil(t, err)
	query, err := parser.ParseQuery("{ as: name }")
	require.Nil(t, err)

	called := false
	observers := &Events{}
	observers.OnField(func(walker *Walker, parentDef *ast.Definition, fieldDef *ast.FieldDefinition, field *ast.Field) {
		called = true

		require.Equal(t, "name", field.Name)
		require.Equal(t, "as", field.Alias)
		require.Equal(t, "Query", parentDef.Name)
	})

	Walk(schema, &query, observers)

	require.True(t, called)
}
