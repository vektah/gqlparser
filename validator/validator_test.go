package validator_test

import (
	"testing"

	"github.com/vektah/gqlparser/v2/validator/rules"

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
	require.NoError(t, err)
	//nolint:staticcheck
	require.Nil(t, validator.Validate(s, q))
	require.Nil(t, validator.ValidateWithRules(s, q, nil))
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
	`,
	})
	require.NoError(t, err)
	//nolint:staticcheck
	r1 := validator.Validate(s, q1)
	require.Len(t, r1, 1)
	const errorString = `SomeOperation:4:2: Field "myAction" argument "myEnum" of type "Locale!" is required, but it was not provided.`
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
	`,
	})
	require.NoError(t, err)
	//nolint:staticcheck
	require.Nil(t, validator.Validate(s, q2))

	// Repeating same query and expecting to still return same validation error
	require.Len(t, r1, 1)
	require.EqualError(t, r1[0], errorString)
}

func TestValidationRulesAreIndependentWithRules(t *testing.T) {
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
	`,
	})
	require.NoError(t, err)
	r1 := validator.ValidateWithRules(s, q1, nil)
	require.Len(t, r1, 1)
	const errorString = `SomeOperation:4:2: Field "myAction" argument "myEnum" of type "Locale!" is required, but it was not provided.`
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
	`,
	})
	require.NoError(t, err)
	require.Nil(t, validator.ValidateWithRules(s, q2, nil))

	// Repeating same query and expecting to still return same validation error
	require.Len(t, r1, 1)
	require.EqualError(t, r1[0], errorString)
}

func TestDeprecatingTypes(t *testing.T) {
	schema := &ast.Source{
		Name: "graph/schema.graphqls",
		Input: `
			type DeprecatedType {
				deprecatedField: String @deprecated
				newField(deprecatedArg: Int): Boolean
			}

			enum DeprecatedEnum {
				ALPHA @deprecated
			}
		`,
		BuiltIn: false,
	}

	_, err := validator.LoadSchema(append([]*ast.Source{validator.Prelude}, schema)...)
	require.NoError(t, err)
}

func TestNoUnusedVariables(t *testing.T) {
	// https://github.com/99designs/gqlgen/issues/2028
	t.Run("gqlgen issues #2028", func(t *testing.T) {
		s := gqlparser.MustLoadSchema(
			&ast.Source{Name: "graph/schema.graphqls", Input: `
	type Query {
		bar: String!
	}
	`, BuiltIn: false},
		)

		q, err := parser.ParseQuery(&ast.Source{Name: "2028", Input: `
			query Foo($flag: Boolean!) {
				...Bar
			}
			fragment Bar on Query {
				bar @include(if: $flag)
			}
		`})
		require.NoError(t, err)
		//nolint:staticcheck
		require.Nil(t, validator.Validate(s, q))
	})
}

func TestNoUnusedVariablesWithRules(t *testing.T) {
	// https://github.com/99designs/gqlgen/issues/2028
	t.Run("gqlgen issues #2028", func(t *testing.T) {
		s := gqlparser.MustLoadSchema(
			&ast.Source{Name: "graph/schema.graphqls", Input: `
	type Query {
		bar: String!
	}
	`, BuiltIn: false},
		)

		q, err := parser.ParseQuery(&ast.Source{Name: "2028", Input: `
			query Foo($flag: Boolean!) {
				...Bar
			}
			fragment Bar on Query {
				bar @include(if: $flag)
			}
		`})
		require.NoError(t, err)
		require.Nil(t, validator.ValidateWithRules(s, q, nil))
	})
}

func TestCustomRuleSet(t *testing.T) {
	someRule := validator.Rule{
		Name: "SomeRule",
		RuleFunc: func(observers *validator.Events, addError validator.AddErrFunc) {
			addError(validator.Message("%s", "some error message"))
		},
	}

	someOtherRule := validator.Rule{
		Name: "SomeOtherRule",
		RuleFunc: func(observers *validator.Events, addError validator.AddErrFunc) {
			addError(validator.Message("%s", "some other error message"))
		},
	}

	s := gqlparser.MustLoadSchema(
		&ast.Source{
			Name: "graph/schema.graphqls",
			Input: `
	type Query {
		bar: String!
	}
	`, BuiltIn: false},
	)

	q, err := parser.ParseQuery(&ast.Source{
		Name: "SomeQuery",
		Input: `
			query Foo($flag: Boolean!) {
				...Bar
			}
		`})
	require.NoError(t, err)
	//nolint:staticcheck
	errList := validator.Validate(s, q, []validator.Rule{someRule, someOtherRule}...)
	require.Len(t, errList, 2)
	require.Equal(t, "some error message", errList[0].Message)
	require.Equal(t, "some other error message", errList[1].Message)
}

func TestCustomRuleSetWithRules(t *testing.T) {
	someRule := validator.Rule{
		Name: "SomeRule",
		RuleFunc: func(observers *validator.Events, addError validator.AddErrFunc) {
			addError(validator.Message("%s", "some error message"))
		},
	}

	someOtherRule := validator.Rule{
		Name: "SomeOtherRule",
		RuleFunc: func(observers *validator.Events, addError validator.AddErrFunc) {
			addError(validator.Message("%s", "some other error message"))
		},
	}

	s := gqlparser.MustLoadSchema(
		&ast.Source{
			Name: "graph/schema.graphqls",
			Input: `
	type Query {
		bar: String!
	}
	`, BuiltIn: false},
	)

	q, err := parser.ParseQuery(&ast.Source{
		Name: "SomeQuery",
		Input: `
			query Foo($flag: Boolean!) {
				...Bar
			}
		`})
	require.NoError(t, err)
	errList := validator.ValidateWithRules(s, q, rules.NewRules(someRule, someOtherRule))
	require.Len(t, errList, 2)

	// because we hold rules in a map, the order is not guaranteed
	// this is fine because we used to add the rule in the init function, so it didn't need to be specified as a requirement for the order.
	messages := []string{errList[0].Message, errList[1].Message}
	require.Contains(t, messages, "some error message")
	require.Contains(t, messages, "some other error message")
}

func TestRemoveRule(t *testing.T) {
	// no error
	validator.RemoveRule("rule that does not exist")

	validator.AddRule("Rule that should no longer exist", func(observers *validator.Events, addError validator.AddErrFunc) {})

	// no error
	validator.RemoveRule("Rule that should no longer exist")
}

func TestValidateWithRulesAndMaximumErrors(t *testing.T) {
	t.Run("maximumErrors limits error count", func(t *testing.T) {
		s := gqlparser.MustLoadSchema(
			&ast.Source{
				Name: "graph/schema.graphqls",
				Input: `
			type Query {
				field1: String!
				field2: String!
				field3: String!
			}
		`, BuiltIn: false},
		)

		q, err := parser.ParseQuery(&ast.Source{
			Name: "SomeQuery",
			Input: `
			query {
				field1
				field2
				field3
			}
		`})
		require.NoError(t, err)

		// Create a rule that generates errors for each field
		errorRule := validator.Rule{
			Name: "ErrorRule",
			RuleFunc: func(observers *validator.Events, addError validator.AddErrFunc) {
				observers.OnField(func(walker *validator.Walker, field *ast.Field) {
					addError(validator.Message("Error for field %s", field.Name))
				})
			},
		}

		rules := rules.NewRules(errorRule)
		errList := validator.ValidateWithRulesAndMaximumErrors(s, q, rules, 2)

		// Should only return 2 errors even though there are 3 fields
		require.Len(t, errList, 2)
	})

	t.Run("maximumErrors zero means no limit", func(t *testing.T) {
		s := gqlparser.MustLoadSchema(
			&ast.Source{
				Name: "graph/schema.graphqls",
				Input: `
			type Query {
				field1: String!
				field2: String!
				field3: String!
			}
		`, BuiltIn: false},
		)

		q, err := parser.ParseQuery(&ast.Source{
			Name: "SomeQuery",
			Input: `
			query {
				field1
				field2
				field3
			}
		`})
		require.NoError(t, err)

		// Create a rule that generates errors for each field
		errorRule := validator.Rule{
			Name: "ErrorRule",
			RuleFunc: func(observers *validator.Events, addError validator.AddErrFunc) {
				observers.OnField(func(walker *validator.Walker, field *ast.Field) {
					addError(validator.Message("Error for field %s", field.Name))
				})
			},
		}

		rules := rules.NewRules(errorRule)
		errList := validator.ValidateWithRulesAndMaximumErrors(s, q, rules, 0)

		// Should return all errors when maximumErrors is 0
		require.Len(t, errList, 3)
	})

	t.Run("negative maximumErrors returns error", func(t *testing.T) {
		s := gqlparser.MustLoadSchema(
			&ast.Source{
				Name: "graph/schema.graphqls",
				Input: `
			type Query {
				field1: String!
			}
		`, BuiltIn: false},
		)

		q, err := parser.ParseQuery(&ast.Source{
			Name: "SomeQuery",
			Input: `
			query {
				field1
			}
		`})
		require.NoError(t, err)

		errList := validator.ValidateWithRulesAndMaximumErrors(s, q, nil, -1)

		// Should return an error about negative maximumErrors
		require.Len(t, errList, 1)
		require.Contains(t, errList[0].Message, "maximumErrors cannot be negative")
	})

	t.Run("maximumErrors stops traversal early", func(t *testing.T) {
		s := gqlparser.MustLoadSchema(
			&ast.Source{
				Name: "graph/schema.graphqls",
				Input: `
			type Query {
				field1: String!
				field2: String!
				field3: String!
				field4: String!
				field5: String!
			}
		`, BuiltIn: false},
		)

		q, err := parser.ParseQuery(&ast.Source{
			Name: "SomeQuery",
			Input: `
			query {
				field1
				field2
				field3
				field4
				field5
			}
		`})
		require.NoError(t, err)

		fieldCount := 0
		// Create a rule that generates errors and counts fields
		errorRule := validator.Rule{
			Name: "ErrorRule",
			RuleFunc: func(observers *validator.Events, addError validator.AddErrFunc) {
				observers.OnField(func(walker *validator.Walker, field *ast.Field) {
					fieldCount++
					addError(validator.Message("Error for field %s", field.Name))
				})
			},
		}

		rules := rules.NewRules(errorRule)
		errList := validator.ValidateWithRulesAndMaximumErrors(s, q, rules, 2)

		// Should only return 2 errors
		require.Len(t, errList, 2)
		// Should have stopped traversal early after exactly 2 fields processed
		require.Equal(t, 2, fieldCount)
	})

	t.Run("maximumErrors with multiple rules", func(t *testing.T) {
		s := gqlparser.MustLoadSchema(
			&ast.Source{
				Name: "graph/schema.graphqls",
				Input: `
			type Query {
				field1: String!
				field2: String!
			}
		`, BuiltIn: false},
		)

		q, err := parser.ParseQuery(&ast.Source{
			Name: "SomeQuery",
			Input: `
			query {
				field1
				field2
			}
		`})
		require.NoError(t, err)

		// Create two rules that each generate errors and count fields
		fieldCount := 0
		rule1 := validator.Rule{
			Name: "Rule1",
			RuleFunc: func(observers *validator.Events, addError validator.AddErrFunc) {
				observers.OnField(func(walker *validator.Walker, field *ast.Field) {
					fieldCount++
					addError(validator.Message("Rule1 error for field %s", field.Name))
				})
			},
		}
		rule2 := validator.Rule{
			Name: "Rule2",
			RuleFunc: func(observers *validator.Events, addError validator.AddErrFunc) {
				observers.OnField(func(walker *validator.Walker, field *ast.Field) {
					fieldCount++
					addError(validator.Message("Rule2 error for field %s", field.Name))
				})
			},
		}

		rules := rules.NewRules(rule1, rule2)
		errList := validator.ValidateWithRulesAndMaximumErrors(s, q, rules, 3)

		// Although we set maximumErrors to 3, we expect 4 errors here (2 rules × 2 fields).
		// The limit is evaluated after the batch is processed, allowing a final overflow.
		require.Equal(t, 4, fieldCount)
		require.Equal(t, 4, len(errList))
	})
}
