package rules_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
	"github.com/vektah/gqlparser/v2/validator/rules"
)

// Each fragment spreads the next one twice. Without memoization the rule
// walks 2^n paths, so a document of about a kilobyte takes seconds to
// validate, and the rule runs whether or not the server serves
// introspection.
func TestMaxIntrospectionDepthIsLinearInFragmentSpreads(t *testing.T) {
	const n = 40
	s := gqlparser.MustLoadSchema(
		&ast.Source{Name: "schema.graphqls", Input: `type Query { a: Int }`},
	)

	var b strings.Builder
	b.WriteString("{ __schema { types { ...F0 } } }\n")
	for i := range n {
		fmt.Fprintf(&b, "fragment F%d on __Type { name ...F%d ...F%d }\n", i, i+1, i+1)
	}
	fmt.Fprintf(&b, "fragment F%d on __Type { name }\n", n)
	q, err := parser.ParseQuery(&ast.Source{Name: "q", Input: b.String()})
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() {
		errs := validator.ValidateWithRules(s, q, rules.NewRules(rules.MaxIntrospectionDepth))
		if len(errs) > 0 {
			done <- errs
			return
		}
		done <- nil
	}()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatalf("validating %d fragments did not finish in 10s", n+1)
	}
}

// Memoizing a fragment must not hide depth: the same fragment reached at a
// greater depth is checked again.
func TestMaxIntrospectionDepthMemoKeepsDepth(t *testing.T) {
	s := gqlparser.MustLoadSchema(
		&ast.Source{Name: "schema.graphqls", Input: `type Query { a: Int }`},
	)
	q, err := parser.ParseQuery(&ast.Source{Name: "q", Input: `
		{ __schema { types { ...F fields { type { fields { type { ...F } } } } } } }
		fragment F on __Type { fields { name } }
	`})
	require.NoError(t, err)
	errs := validator.ValidateWithRules(s, q, rules.NewRules(rules.MaxIntrospectionDepth))
	require.Len(t, errs, 1)
	require.Equal(t, "Maximum introspection depth exceeded", errs[0].Message)
}
