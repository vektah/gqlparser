package validator

import (
	"os"
	"testing"

	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser/testrunner"
)

func TestLoadSchema(t *testing.T) {
	t.Run("prelude", func(t *testing.T) {
		s, err := LoadSchema(Prelude)
		require.NoError(t, err)

		boolDef := s.Types["Boolean"]
		require.Equal(t, "Boolean", boolDef.Name)
		require.Equal(t, ast.Scalar, boolDef.Kind)
		require.Equal(t, "The `Boolean` scalar type represents `true` or `false`.", boolDef.Description)

		deferDef := s.Directives["defer"]
		require.Equal(t, "defer", deferDef.Name, "@defer exists.")
		require.Equal(t, "if", deferDef.Arguments[0].Name, "@defer has \"if\" argument.")
		require.Equal(t, "label", deferDef.Arguments[1].Name, "@defer has \"label\" argument.")
	})
	t.Run("swapi", func(t *testing.T) {
		file, err := os.ReadFile("testdata/swapi.graphql")
		require.NoError(t, err)
		s, err := LoadSchema(Prelude, &ast.Source{Input: string(file), Name: "TestLoadSchema"})
		require.NoError(t, err)

		require.Equal(t, "Query", s.Query.Name)
		require.Equal(t, "hero", s.Query.Fields[0].Name)

		require.Equal(t, "Human", s.Types["Human"].Name)

		require.Equal(t, "Subscription", s.Subscription.Name)
		require.Equal(t, "reviewAdded", s.Subscription.Fields[0].Name)

		possibleCharacters := s.GetPossibleTypes(s.Types["Character"])
		require.Len(t, possibleCharacters, 2)
		require.Equal(t, "Human", possibleCharacters[0].Name)
		require.Equal(t, "Droid", possibleCharacters[1].Name)

		implements := s.GetImplements(s.Types["Droid"])
		require.Len(t, implements, 2)
		require.Equal(t, "Character", implements[0].Name)    // interface
		require.Equal(t, "SearchResult", implements[1].Name) // union
	})

	t.Run("default root operation type names", func(t *testing.T) {
		file, err := os.ReadFile("testdata/default_root_operation_type_names.graphql")
		require.NoError(t, err)
		s, err := LoadSchema(Prelude, &ast.Source{Input: string(file), Name: "TestLoadSchema"})
		require.NoError(t, err)

		require.Nil(t, s.Mutation)
		require.Nil(t, s.Subscription)

		require.Equal(t, "Mutation", s.Types["Mutation"].Name)
		require.Equal(t, "Subscription", s.Types["Subscription"].Name)
	})

	t.Run("type extensions", func(t *testing.T) {
		file, err := os.ReadFile("testdata/extensions.graphql")
		require.NoError(t, err)
		s, err := LoadSchema(Prelude, &ast.Source{Input: string(file), Name: "TestLoadSchema"})
		require.NoError(t, err)

		require.Equal(t, "Subscription", s.Subscription.Name)
		require.Equal(t, "dogEvents", s.Subscription.Fields[0].Name)

		require.Equal(t, 1, len(s.SchemaDirectives))
		require.Equal(t, "exampleOnSchemaDirective", s.SchemaDirectives[0].Name)
		require.Equal(t, 1, len(s.SchemaDirectives[0].Arguments))
		require.Equal(t, "name", s.SchemaDirectives[0].Arguments[0].Name)
		require.Equal(t, "foo", s.SchemaDirectives[0].Arguments[0].Value.Raw)

		require.Equal(t, "owner", s.Types["Dog"].Fields[1].Name)

		directives := s.Types["Person"].Directives
		require.Len(t, directives, 2)
		wantArgs := []string{"sushi", "tempura"}
		for i, directive := range directives {
			require.Equal(t, "favorite", directive.Name)
			require.True(t, directive.Definition.IsRepeatable)
			for _, arg := range directive.Arguments {
				require.Equal(t, wantArgs[i], arg.Value.Raw)
			}
		}
	})

	t.Run("interfaces", func(t *testing.T) {
		file, err := os.ReadFile("testdata/interfaces.graphql")
		require.NoError(t, err)
		s, err := LoadSchema(Prelude, &ast.Source{Input: string(file), Name: "interfaces"})
		require.NoError(t, err)

		implements := s.GetImplements(s.Types["Canine"])
		require.Len(t, implements, 1)
		require.Equal(t, "Mammal", implements[0].Name)

		possibleTypes := s.GetPossibleTypes(s.Types["Mammal"])
		require.Len(t, possibleTypes, 1)
		require.Equal(t, "Canine", possibleTypes[0].Name)
	})

	testrunner.Test(t, "./schema_test.yml", func(t *testing.T, input string) testrunner.Spec {
		_, err := LoadSchema(Prelude, &ast.Source{Input: input})
		if err != nil {
			return testrunner.Spec{
				Error: err.(*gqlerror.Error),
			}
		}
		return testrunner.Spec{}
	})
}

func TestSchemaDescription(t *testing.T) {
	s, err := LoadSchema(Prelude, &ast.Source{Name: "graph/schema.graphqls", Input: `
	"""
	A simple GraphQL schema which is well described.
	"""
	schema {
		query: Query
	}

	type Query {
		entity: String
	}
	`, BuiltIn: false})
	require.NoError(t, err)
	want := "A simple GraphQL schema which is well described."
	require.Equal(t, want, s.Description)
}

func TestSchemaDescriptionWithQuotesAtEnd(t *testing.T) {
	// This test demonstrates a bug in the parser where quotes at the end of a
	// description without a space cause parsing errors

	t.Run("working case - quotes followed by space at end of description", func(t *testing.T) {
		// This case works correctly - note the space after the quote and before the closing """
		_, err := LoadSchema(Prelude, &ast.Source{Name: "test", Input: `
		"""This is a "test" """
		type Query {
		  field: String
		}
		`, BuiltIn: false})
		require.NoError(t, err, "Schema with quotes followed by space at end of description should parse successfully")
	})

	t.Run("bug - quotes at end of description", func(t *testing.T) {
		// This case fails - note the quote directly before the closing """
		_, err := LoadSchema(Prelude, &ast.Source{Name: "test", Input: `
		"""This is a "test""""
		type Query {
		  field: String
		}
		`, BuiltIn: false})
		require.NoError(t, err, "Schema with quotes at end of description should parse successfully")
	})
}
