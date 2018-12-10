package fieldmap

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/ast"
)

func TestPath(t *testing.T) {
	require.True(t, Path{
		&ast.Field{Alias: "a"},
		&ast.Field{Alias: "b"},
	}.Equal(Path{
		&ast.Field{Alias: "a"},
		&ast.Field{Alias: "b"},
	}))

	require.False(t, Path{
		&ast.Field{Alias: "a"},
		&ast.Field{Alias: "c"},
	}.Equal(Path{
		&ast.Field{Alias: "a"},
		&ast.Field{Alias: "b"},
	}))

	require.True(t, Path{
		&ast.Field{Alias: "a"},
		&ast.InlineFragment{TypeCondition: ""},
		&ast.Field{Alias: "b"},
	}.Equal(Path{
		&ast.Field{Alias: "a"},
		&ast.Field{Alias: "b"},
	}))

	require.True(t, Path{
		&ast.Field{Alias: "a"},
		&ast.Field{Alias: "b"},
	}.Equal(Path{
		&ast.Field{Alias: "a"},
		&ast.InlineFragment{TypeCondition: ""},
		&ast.Field{Alias: "b"},
	}))

	require.True(t, Path{
		&ast.Field{Alias: "a"},
		&ast.InlineFragment{TypeCondition: "dog"},
		&ast.Field{Alias: "b"},
	}.Equal(Path{
		&ast.Field{Alias: "a"},
		&ast.InlineFragment{TypeCondition: "dog"},
		&ast.Field{Alias: "b"},
	}))

	require.False(t, Path{
		&ast.Field{Alias: "a"},
		&ast.InlineFragment{TypeCondition: "cat"},
		&ast.Field{Alias: "b"},
	}.Equal(Path{
		&ast.Field{Alias: "a"},
		&ast.InlineFragment{TypeCondition: "dog"},
		&ast.Field{Alias: "b"},
	}))
}
