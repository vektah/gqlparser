package validator_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
)

func TestExtendingNonExistantTypes(t *testing.T) {
	s := gqlparser.MustLoadSchema(
		&ast.Source{Name: "graph/schema.graphqls", Input: `
extend type User {
    id: ID!
}

extend type Product {
    upc: String!
}

union _Entity = Product | User

extend type Query {
	entity: _Entity
}
`, BuiltIn: false},
	)

	q, err := parser.ParseQuery(&ast.Source{Name: "ff", Input: `{
		entity {
		  ... on User {
			id
		  }
		}
	}`})
	require.Nil(t, err)
	require.Nil(t, validator.Validate(s, q))
}

func TestValidationRulesAreIndependent(t *testing.T) {
	s := gqlparser.MustLoadSchema(
		&ast.Source{Name: "graph/schema.graphqls", Input: `
extend type Query {
    myAction(myEnum: Locale!): SomeResult!
}

type SomeResult {
    id: String
}

enum Locale {
    EN
    LT
    DE
}
`, BuiltIn: false},
	)

	// Validation as a first call
	q1, err := parser.ParseQuery(&ast.Source{
		Name: "SomeOperation", Input: `
query SomeOperation {
	# Note: Not providing mandatory parameter: (myEnum: Locale!)
	myAction {
		id
	}
}
	`})
	require.Nil(t, err)
	r1 := validator.Validate(s, q1)
	require.Len(t, r1, 1)
	const errorString = `SomeOperation:4: Field "myAction" argument "myEnum" of type "Locale!" is required but not provided.`
	require.EqualError(t, r1[0], errorString)

	// Some other call that should not affect validator behavior
	q2, err := parser.ParseQuery(&ast.Source{
		Name: "SomeOperation", Input: `
# Note: there is default enum value in variables
query SomeOperation ($locale: Locale! = DE) {
	myAction(myEnum: $locale) {
		id
	}
}
	`})
	require.Nil(t, err)
	require.Nil(t, validator.Validate(s, q2))

	// Repeating same query and expecting to still return same validation error
	require.Len(t, r1, 1)
	require.EqualError(t, r1[0], errorString)
}
