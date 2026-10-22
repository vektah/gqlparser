package validator_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator"
)

func TestPreludeDocument(t *testing.T) {
	t.Run("every declaration is built in", func(t *testing.T) {
		doc := validator.PreludeDocument()

		require.NotEmpty(t, doc.Directives)
		for _, dir := range doc.Directives {
			require.True(t, dir.Position.Src.BuiltIn, "directive @%s", dir.Name)
		}
		require.NotEmpty(t, doc.Definitions)
		for _, def := range doc.Definitions {
			require.True(t, def.BuiltIn, "type %s", def.Name)
		}
	})

	t.Run("each call returns an independent copy", func(t *testing.T) {
		first := validator.PreludeDocument()
		first.Directives.ForName("skip").Name = "changed"
		first.Definitions = nil

		second := validator.PreludeDocument()
		require.NotNil(t, second.Directives.ForName("skip"))
		require.NotNil(t, second.Definitions.ForName("__Schema"))
	})
}

func TestQueryIntrospectionFields(t *testing.T) {
	t.Run("LoadSchema adds them to the query type", func(t *testing.T) {
		s, err := validator.LoadSchema(
			validator.Prelude,
			&ast.Source{Input: "type Query { f: Int }"},
		)
		require.NoError(t, err)

		want := validator.QueryIntrospectionFields()
		require.Equal(t, want, s.Query.Fields[len(s.Query.Fields)-len(want):])
	})

	t.Run("each call returns new definitions", func(t *testing.T) {
		first := validator.QueryIntrospectionFields()
		first.ForName("__type").Arguments = nil

		second := validator.QueryIntrospectionFields()
		require.NotNil(t, second.ForName("__type").Arguments.ForName("name"))
	})
}
