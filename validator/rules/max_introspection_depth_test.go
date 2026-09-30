package rules_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
	"github.com/vektah/gqlparser/v2/validator/rules"
)

// timeMultiplier scales the budget in TestMaxIntrospectionDepthIsLinearInFragmentSpreads.
// CI can increase this.
var timeMultiplier = time.Duration(1)

// cyclicQuery reaches F1 at one depth twice: once below F0, where the cycle back to F0 is
// cut, and once beside it, where following F0 would reach the depth limit. Both tests below
// have to describe the same document, or neither says anything about the other.
const cyclicQuery = `
	{ __schema { types { ...F0 fields { type { ...F1 } } } } }
	fragment F0 on __Type { fields { type { ...F1 } } }
	fragment F1 on __Type { fields { type { ...F0 } } }
`

const introspectionTestSchema = `type Query { a: Int }`

func TestMaxIntrospectionDepth(t *testing.T) {
	cases := map[string]struct {
		query        string
		wantMessages []string
	}{
		// Memoizing a fragment must not hide depth: the same fragment reached at a greater
		// depth is checked again.
		"fragment reached again at a greater depth": {
			query: `
				{ __schema { types { ...F fields { type { fields { type { ...F } } } } } } }
				fragment F on __Type { fields { name } }
			`,
			wantMessages: []string{"Maximum introspection depth exceeded"},
		},
		// This case records what the rule does today, not what it ought to do. The memo
		// answers for F1 with a result that was cut short under a different set of visited
		// fragments, so the depth error the uncached walk reports is missed. See
		// depthChecker for why that is accepted, and
		// TestMaxIntrospectionDepthCyclicDocumentStillFailsValidation for what rejects the
		// document instead. Do not "fix" this by expecting an error: keying the memo so
		// that it answers exactly restores an exponential walk for cyclic documents.
		"cyclic document, memo answers before the limit is reached": {
			query:        cyclicQuery,
			wantMessages: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s := gqlparser.MustLoadSchema(
				&ast.Source{Name: "schema.graphqls", Input: introspectionTestSchema},
			)
			q, err := parser.ParseQuery(&ast.Source{Name: "q", Input: tc.query})
			require.NoError(t, err)

			errs := validator.ValidateWithRules(s, q, rules.NewRules(rules.MaxIntrospectionDepth))

			var messages []string
			for _, e := range errs {
				messages = append(messages, e.Message)
			}
			require.Equal(t, tc.wantMessages, messages)
		})
	}
}

// The depth rule missing an error on a cyclic document is only acceptable because such a
// document fails validation anyway. If that stops holding, depthChecker's memo has to be
// keyed differently, whatever that costs.
func TestMaxIntrospectionDepthCyclicDocumentStillFailsValidation(t *testing.T) {
	s := gqlparser.MustLoadSchema(
		&ast.Source{Name: "schema.graphqls", Input: introspectionTestSchema},
	)
	q, err := parser.ParseQuery(&ast.Source{Name: "q", Input: cyclicQuery})
	require.NoError(t, err)

	errs := validator.ValidateWithRules(s, q, rules.NewDefaultRules())

	require.Len(t, errs, 1)
	require.Equal(t, "NoFragmentCycles", errs[0].Rule)
}

// Each fragment spreads the next one twice. Following every path visits 2^n of them, so
// before this was memoized a document of about a kilobyte took seconds to validate, and the
// rule runs whether or not the server serves introspection.
//
// The budget is generous on purpose: memoized, this finishes in under a millisecond, and a
// document of 41 fragments walked exhaustively takes hours. Anything in between is a
// regression, so there is no need to time it finely.
func TestMaxIntrospectionDepthIsLinearInFragmentSpreads(t *testing.T) {
	const n = 40
	s := gqlparser.MustLoadSchema(
		&ast.Source{Name: "schema.graphqls", Input: introspectionTestSchema},
	)

	var b strings.Builder
	b.WriteString("{ __schema { types { ...F0 } } }\n")
	for i := range n {
		fmt.Fprintf(&b, "fragment F%d on __Type { name ...F%d ...F%d }\n", i, i+1, i+1)
	}
	fmt.Fprintf(&b, "fragment F%d on __Type { name }\n", n)
	q, err := parser.ParseQuery(&ast.Source{Name: "q", Input: b.String()})
	require.NoError(t, err)

	// The walk cannot be cancelled, so on failure this goroutine keeps running until the
	// test binary exits. That is the price of not waiting hours for it.
	done := make(chan gqlerror.List, 1)
	go func() {
		done <- validator.ValidateWithRules(s, q, rules.NewRules(rules.MaxIntrospectionDepth))
	}()

	budget := 20 * time.Second * timeMultiplier
	select {
	case errs := <-done:
		require.Empty(t, errs)
	case <-time.After(budget):
		t.Fatalf("validating %d fragments did not finish in %s", n+1, budget)
	}
}
