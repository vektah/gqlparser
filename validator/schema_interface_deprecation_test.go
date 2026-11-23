package validator_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestLoadSchemaInterfaceFieldDeprecation(t *testing.T) {
	for _, tc := range []struct {
		name, sdl, err string
	}{
		{"object", `interface Node { id: ID! } type Item implements Node { id: ID! @deprecated }`, "Item.id"},
		{"interface", `interface Node { id: ID! } interface Child implements Node { id: ID! @deprecated }`, "Child.id"},
		{"both deprecated", `interface Node { id: ID! @deprecated } type Item implements Node { id: ID! @deprecated(reason: "Use key.") }`, ""},
		{"interface only deprecated", `interface Node { id: ID! @deprecated } type Item implements Node { id: ID! }`, ""},
		{"object only field", `interface Node { id: ID! } type Item implements Node { id: ID! oldId: ID @deprecated }`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := gqlparser.LoadSchema(&ast.Source{Name: "schema.graphql", Input: tc.sdl})
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.err)
			}
		})
	}
}
