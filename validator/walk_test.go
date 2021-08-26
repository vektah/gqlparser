package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

func TestWalker(t *testing.T) {
	schema, err := LoadSchema(Prelude, &ast.Source{Input: "type Query { name: String }\n schema { query: Query }"})
	require.Nil(t, err)
	query, err := parser.ParseQuery(&ast.Source{Input: "{ as: name }"})
	require.Nil(t, err)

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
	require.Nil(t, err)
	query, err := parser.ParseQuery(&ast.Source{Input: "{ ... { name } }"})
	require.Nil(t, err)

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

func TestShallowCopy(t *testing.T) {
	input := &ast.Type{
		NamedType: "Some",
		Elem:      nil,
		NonNull:   false,
		Position: &ast.Position{
			Start:  10,
			End:    20,
			Line:   3,
			Column: 4,
			Src:    nil,
		},
	}
	copied := shallowCopy(input)
	require.Equal(t, input, copied, "have same values")
	require.NotEqual(t, fmt.Sprintf("%p", input), fmt.Sprintf("%p", copied), "different pointers")
	copied.NonNull = true
	require.Equal(t, false, input.NonNull, "not change by reference")
	require.Equal(t, true, copied.NonNull, "but changed on a copy")
}
