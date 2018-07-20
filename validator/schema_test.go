package validator

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/parser/testrunner"
)

func TestLoadSchema(t *testing.T) {
	t.Run("swapi", func(t *testing.T) {
		file, err := ioutil.ReadFile("testdata/swapi.graphql")
		require.Nil(t, err)
		s, err := LoadSchema(Prelude, &ast.Source{Input: string(file), Name: "TestLoadSchema"})
		require.Nil(t, err)

		require.Equal(t, "Query", s.Query.Name)
		require.Equal(t, "hero", s.Query.Fields[0].Name)

		require.Equal(t, "Human", s.Types["Human"].Name)

		require.Equal(t, "Subscription", s.Subscription.Name)
		require.Equal(t, "reviewAdded", s.Subscription.Fields[0].Name)

		possibleCharacters := s.GetPossibleTypes(s.Types["Character"])
		require.Len(t, possibleCharacters, 2)
		require.Equal(t, "Human", possibleCharacters[0].Name)
		require.Equal(t, "Droid", possibleCharacters[1].Name)
	})

	t.Run("type extensions", func(t *testing.T) {
		file, err := ioutil.ReadFile("testdata/extensions.graphql")
		require.Nil(t, err)
		s, err := LoadSchema(Prelude, &ast.Source{Input: string(file), Name: "TestLoadSchema"})
		require.Nil(t, err)

		require.Equal(t, "Subscription", s.Subscription.Name)
		require.Equal(t, "dogEvents", s.Subscription.Fields[0].Name)

		require.Equal(t, "owner", s.Types["Dog"].Fields[1].Name)
	})

	testrunner.Test(t, "./schema_test.yml", func(t *testing.T, input string) testrunner.Spec {
		_, err := LoadSchema(Prelude, &ast.Source{Input: input})
		return testrunner.Spec{
			Error: err,
		}
	})
}
