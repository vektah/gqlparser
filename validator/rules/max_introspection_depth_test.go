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

// cyclicQuery's fragments spread one another through a list field, so each lap of the cycle
// nests one level deeper. Both tests below have to describe the same document, or neither
// says anything about the other.
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
		// One checker serves the whole document, so these two cases check that sharing a
		// fragment's result between introspection roots does not carry the wrong answer
		// across them: F is under the limit from the shallow root and over it from the
		// deep one, in either order.
		"same fragment from two roots, shallow root first": {
			query: `
				{
					shallow: __schema { types { ...F } }
					deep: __schema { types { fields { type { fields { type { ...F } } } } } }
				}
				fragment F on __Type { fields { name } }
			`,
			wantMessages: []string{"Maximum introspection depth exceeded"},
		},
		"same fragment from two roots, deep root first": {
			query: `
				{
					deep: __schema { types { fields { type { fields { type { ...F } } } } } }
					shallow: __schema { types { ...F } }
				}
				fragment F on __Type { fields { name } }
			`,
			wantMessages: []string{"Maximum introspection depth exceeded"},
		},
		// Records what the rule does today, not what it ought to do. One checker serves the
		// whole document, so F, answered from G's entry depth while G was being walked, is
		// reused by the root that spreads F directly and would have walked it. An uncached
		// walk reports both roots; this reports the first. depthChecker says why that is
		// worth the memo. Do not "fix" it by dropping the shared checker: a cut only
		// happens on a cycle, so no valid document is affected.
		"cycle reached from two roots reports the first": {
			query: `
				{
					a: __schema { types { ...G } }
					b: __schema { types { ...F } }
				}
				fragment G on __Type { ...F fields { fields { fields { name } } } }
				fragment F on __Type { ...G }
			`,
			wantMessages: []string{"Maximum introspection depth exceeded"},
		},
		// A lap of this cycle nests one level deeper, so repeating it passes any limit.
		"cycle that gains depth on each lap": {
			query:        cyclicQuery,
			wantMessages: []string{"Maximum introspection depth exceeded"},
		},
		// The boundary of that rule, and the one cyclic case the graphql-js suite pins
		// (MaxIntrospectionDepthRule.spec.yml, "doesn't infinitely recurse on fragment
		// cycle"): a lap that gains no depth reaches nothing a further lap would, so it is
		// cut and reports nothing.
		"cycle that gains no depth": {
			query: `
				{ __schema { types { ...Cycle } } }
				fragment Cycle on __Type { ...Cycle }
			`,
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

// Under the default rules the document is rejected twice over, by this rule and by the one
// that owns fragment cycles. The rule no longer depends on that second error.
func TestMaxIntrospectionDepthCyclicDocumentStillFailsValidation(t *testing.T) {
	s := gqlparser.MustLoadSchema(
		&ast.Source{Name: "schema.graphqls", Input: introspectionTestSchema},
	)
	q, err := parser.ParseQuery(&ast.Source{Name: "q", Input: cyclicQuery})
	require.NoError(t, err)

	errs := validator.ValidateWithRules(s, q, rules.NewDefaultRules())

	rulesHit := make([]string, 0, len(errs))
	for _, e := range errs {
		rulesHit = append(rulesHit, e.Rule)
	}
	require.ElementsMatch(t, []string{"MaxIntrospectionDepth", "NoFragmentCycles"}, rulesHit)
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
