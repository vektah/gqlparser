package validator_test

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
	"github.com/vektah/gqlparser/v2/validator/core"
	"github.com/vektah/gqlparser/v2/validator/rules"
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

	require.Nil(t, validator.Validate(s, q))
	require.Nil(t, validator.ValidateWithRules(s, q, nil))
}

func TestVariablesInAllowedPositionRetainsSourcesForMultipleLocations(t *testing.T) {
	s := gqlparser.MustLoadSchema(&ast.Source{
		Name: "schema.graphqls",
		Input: `
input Filter @oneOf {
  byID: ID
}

type Query {
  search(filter: Filter): String
}
`,
	})

	querySource := &ast.Source{
		Name:  "queries/Search.graphql",
		Input: "query Search($id: ID) { ...SearchFields }",
	}
	fragmentSource := &ast.Source{
		Name:  "fragments/SearchFields.graphql",
		Input: "fragment SearchFields on Query { search(filter: {byID: $id}) }",
	}
	query, err := parser.ParseQuery(querySource)
	require.NoError(t, err)
	fragment, err := parser.ParseQuery(fragmentSource)
	require.NoError(t, err)

	doc := &ast.QueryDocument{
		Operations: query.Operations,
		Fragments:  fragment.Fragments,
	}
	errs := validator.ValidateWithSources(s, doc)

	var found *gqlerror.ErrorWithSources
	for _, err := range errs {
		if err.Rule == "VariablesInAllowedPosition" {
			found = err
			break
		}
	}
	require.NotNil(t, found, "expected a VariablesInAllowedPosition error, got %v", errs)
	require.Len(t, found.Locations, 2)
	require.Equal(
		t,
		gqlerror.SourceLocation{Line: 1, Column: 14, Source: querySource},
		found.Locations[0],
	)
	require.Equal(
		t,
		gqlerror.SourceLocation{Line: 1, Column: 56, Source: fragmentSource},
		found.Locations[1],
	)
	require.Equal(t, 1, found.Locations[0].Line)
	require.Equal(t, 1, found.Locations[1].Line)
	require.Equal(t, 14, found.Locations[0].Column)
	require.Equal(t, 56, found.Locations[1].Column)
	require.Equal(t, fragmentSource.Name, found.Extensions["file"])
	require.Equal(t, "queries/Search.graphql:1:14: "+found.Message, found.Error())

	encoded, marshalErr := json.Marshal(found)
	require.NoError(t, marshalErr)
	require.JSONEq(t, `{
		"message": "Variable \"$id\" is of type \"ID\" but must be non-nullable to be used for OneOf Input Object \"Filter\".",
		"locations": [
			{"line": 1, "column": 14},
			{"line": 1, "column": 56}
		],
		"extensions": {"file": "fragments/SearchFields.graphql"}
	}`, string(encoded))
}

func TestSingleLocationValidationRetainsSource(t *testing.T) {
	s := gqlparser.MustLoadSchema(&ast.Source{
		Name:  "schema.graphqls",
		Input: "type Query { search: String }",
	})
	source := &ast.Source{
		Name:  "queries/Search.graphql",
		Input: "query Search { missing }",
	}
	query, err := parser.ParseQuery(source)
	require.NoError(t, err)

	var found *gqlerror.ErrorWithSources
	for _, err := range validator.ValidateWithSources(s, query) {
		if err.Rule == "FieldsOnCorrectType" {
			found = err
			break
		}
	}
	require.NotNil(t, found)
	require.Len(t, found.Locations, 1)
	require.Equal(
		t,
		gqlerror.SourceLocation{Line: 1, Column: 16, Source: source},
		found.Locations[0],
	)
}

func TestValidateWithRulesWithSourcesRetainsOrderedCustomRuleLocations(t *testing.T) {
	s := gqlparser.MustLoadSchema(&ast.Source{
		Name:  "schema.graphqls",
		Input: "type Query { field: String }",
	})
	doc, err := parser.ParseQuery(&ast.Source{
		Name:  "query.graphql",
		Input: "query Query { field }",
	})
	require.NoError(t, err)

	firstSource := &ast.Source{Name: "first.graphql"}
	secondSource := &ast.Source{Name: "second.graphql"}
	multiRule := validator.Rule{
		Name: "MultiRule",
		RuleFunc: func(_ *validator.Events, addError validator.AddErrFunc) {
			addError(
				validator.Message("multi"),
				core.At(&ast.Position{Src: firstSource, Line: 2, Column: 3}),
				core.At(&ast.Position{Src: secondSource, Line: 4, Column: 5}),
			)
		},
	}
	alphaRule := validator.Rule{
		Name: "AlphaRule",
		RuleFunc: func(_ *validator.Events, addError validator.AddErrFunc) {
			addError(
				validator.Message("alpha"),
				core.At(&ast.Position{Src: secondSource, Line: 6, Column: 7}),
			)
		},
	}

	for range 5 {
		errs := validator.ValidateWithRulesWithSources(s, doc, rules.NewRules(multiRule, alphaRule))
		require.Len(t, errs, 2)
		require.Equal(t, "AlphaRule", errs[0].Rule)
		require.Equal(t, "MultiRule", errs[1].Rule)
		require.Equal(
			t,
			[]gqlerror.SourceLocation{{Line: 6, Column: 7, Source: secondSource}},
			errs[0].Locations,
		)
		require.Equal(t, []gqlerror.SourceLocation{
			{Line: 2, Column: 3, Source: firstSource},
			{Line: 4, Column: 5, Source: secondSource},
		}, errs[1].Locations)
	}
}

func TestValidateWithRulesWithSourcesIsConcurrent(t *testing.T) {
	const runs = 16
	type result struct {
		source *ast.Source
		line   int
		errs   gqlerror.SourceList
		err    error
	}
	results := make(chan result, runs)
	var wg sync.WaitGroup

	for i := 0; i < runs; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := gqlparser.MustLoadSchema(&ast.Source{
				Name:  "schema.graphqls",
				Input: "type Query { field: String }",
			})
			doc, err := parser.ParseQuery(&ast.Source{
				Name:  "query.graphql",
				Input: "query Query { field }",
			})
			if err != nil {
				results <- result{err: err}
				return
			}
			source := &ast.Source{Name: fmt.Sprintf("query-%d.graphql", i)}
			rule := validator.Rule{
				Name: "ParallelRule",
				RuleFunc: func(_ *validator.Events, addError validator.AddErrFunc) {
					addError(
						validator.Message("parallel"),
						core.At(&ast.Position{Src: source, Line: i + 1, Column: 2}),
					)
				},
			}
			results <- result{
				source: source,
				line:   i + 1,
				errs:   validator.ValidateWithRulesWithSources(s, doc, rules.NewRules(rule)),
			}
		}()
	}

	wg.Wait()
	close(results)
	for result := range results {
		require.NoError(t, result.err)
		require.Len(t, result.errs, 1)
		require.Equal(t, []gqlerror.SourceLocation{{
			Line:   result.line,
			Column: 2,
			Source: result.source,
		}}, result.errs[0].Locations)
	}
}

func TestValidateWithRulesWithSourcesAllowsErrorsWithoutLocations(t *testing.T) {
	s := gqlparser.MustLoadSchema(&ast.Source{
		Name:  "schema.graphqls",
		Input: "type Query { field: String }",
	})
	doc, err := parser.ParseQuery(&ast.Source{
		Name:  "query.graphql",
		Input: "query Query { field }",
	})
	require.NoError(t, err)

	rule := validator.Rule{
		Name: "NoLocationRule",
		RuleFunc: func(_ *validator.Events, addError validator.AddErrFunc) {
			addError(validator.Message("no location"))
		},
	}
	errs := validator.ValidateWithRulesWithSources(s, doc, rules.NewRules(rule))

	require.Len(t, errs, 1)
	require.Equal(t, "no location", errs[0].Message)
	require.Nil(t, errs[0].Locations)
}

func TestValidateWithRulesWithSourcesUsesDefaultRules(t *testing.T) {
	s := gqlparser.MustLoadSchema(&ast.Source{
		Name:  "schema.graphqls",
		Input: "type Query { field: String }",
	})
	source := &ast.Source{
		Name:  "query.graphql",
		Input: "query Query { missing }",
	}
	doc, err := parser.ParseQuery(source)
	require.NoError(t, err)

	var found *gqlerror.ErrorWithSources
	for _, validationErr := range validator.ValidateWithRulesWithSources(s, doc, nil) {
		if validationErr.Rule == "FieldsOnCorrectType" {
			found = validationErr
			break
		}
	}

	require.NotNil(t, found)
	require.Equal(
		t,
		[]gqlerror.SourceLocation{{Line: 1, Column: 15, Source: source}},
		found.Locations,
	)
}

func TestCaptureSourceLocationsRejectsClearedLocations(t *testing.T) {
	err := &gqlerror.Error{}
	source := &ast.Source{Name: "query.graphql"}

	require.PanicsWithValue(
		t,
		"gqlparser: captured source location 0 does not match the final error locations",
		func() {
			core.CaptureSourceLocations(err, func() {
				core.At(&ast.Position{Src: source, Line: 1, Column: 2})(err)
				err.Locations = nil
			})
		},
	)
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

	t.Run("variable used in fragment definition directive", func(t *testing.T) {
		s := gqlparser.MustLoadSchema(
			&ast.Source{Name: "graph/schema.graphqls", Input: `
	directive @testDirective(x: Int) on FRAGMENT_DEFINITION

	type Query {
		bar: String!
	}
	`, BuiltIn: false},
		)

		q, err := parser.ParseQuery(&ast.Source{Name: "fragmentDefinitionDirective", Input: `
			query Foo($x: Int) {
				...Bar
			}

			fragment Bar on Query @testDirective(x: $x) {
				bar
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
	`, BuiltIn: false,
		},
	)

	q, err := parser.ParseQuery(&ast.Source{
		Name: "SomeQuery",
		Input: `
			query Foo($flag: Boolean!) {
				...Bar
			}
		`,
	})
	require.NoError(t, err)

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
	`, BuiltIn: false,
		},
	)

	q, err := parser.ParseQuery(&ast.Source{
		Name: "SomeQuery",
		Input: `
			query Foo($flag: Boolean!) {
				...Bar
			}
		`,
	})
	require.NoError(t, err)
	errList := validator.ValidateWithRules(s, q, rules.NewRules(someRule, someOtherRule))
	require.Len(t, errList, 2)

	// because we hold rules in a map, the order is not guaranteed
	// this is fine because we used to add the rule in the init function, so it didn't need to be
	// specified as a requirement for the order.
	messages := []string{errList[0].Message, errList[1].Message}
	require.Contains(t, messages, "some error message")
	require.Contains(t, messages, "some other error message")
}

func TestRemoveRule(t *testing.T) {
	// no error
	validator.RemoveRule("rule that does not exist")

	validator.AddRule(
		"Rule that should no longer exist",
		func(observers *validator.Events, addError validator.AddErrFunc) {},
	)

	// no error
	validator.RemoveRule("Rule that should no longer exist")
}
