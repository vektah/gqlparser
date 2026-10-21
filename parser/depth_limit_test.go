package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2/ast"
)

// nestedQuery builds a query whose selection sets nest exactly depth levels,
// e.g. depth=3 -> "{a{a{x}}}".
func nestedQuery(depth int) string {
	if depth < 1 {
		return "{x}"
	}
	return "{" + strings.Repeat("a{", depth-1) + "x" + strings.Repeat("}", depth)
}

func TestParseQueryDepthLimit(t *testing.T) {
	t.Run("shallow query parses cleanly", func(t *testing.T) {
		doc, err := ParseQuery(&ast.Source{Name: "q", Input: "{ a { b { c } } }"})
		require.NoError(t, err)
		require.NotNil(t, doc)
	})

	t.Run("legitimately deep query under the default parses", func(t *testing.T) {
		doc, err := ParseQuery(&ast.Source{Name: "q", Input: nestedQuery(DefaultMaxDepthLimit - 1)})
		require.NoError(t, err)
		require.NotNil(t, doc)
	})

	t.Run("deeply nested query errors instead of overflowing the stack", func(t *testing.T) {
		// Before the fix this input drove unbounded recursion to a fatal
		// runtime stack overflow. It must now return a normal error.
		_, err := ParseQuery(&ast.Source{Name: "q", Input: nestedQuery(700000)})
		require.Error(t, err)
		require.Contains(t, err.Error(), "exceeded maximum selection set nesting depth")
	})

	t.Run("boundary: at the limit parses, one over errors", func(t *testing.T) {
		const limit = 32

		atLimit, err := ParseQueryWithLimits(&ast.Source{Name: "q", Input: nestedQuery(limit)}, 0, limit)
		require.NoError(t, err)
		require.NotNil(t, atLimit)

		_, err = ParseQueryWithLimits(&ast.Source{Name: "q", Input: nestedQuery(limit + 1)}, 0, limit)
		require.Error(t, err)
		require.Contains(t, err.Error(), "exceeded maximum selection set nesting depth of 32")
	})

	t.Run("zero depth limit means unlimited (opt-out)", func(t *testing.T) {
		// A modest depth that is safe to parse unbounded; confirms 0 disables
		// the guard rather than rejecting at the default.
		doc, err := ParseQueryWithLimits(&ast.Source{Name: "q", Input: nestedQuery(2000)}, 0, 0)
		require.NoError(t, err)
		require.NotNil(t, doc)
	})

	t.Run("depth resets across sibling selection sets", func(t *testing.T) {
		// Two shallow operations in one document must not accumulate depth.
		input := "query A { a { b } } query B { c { d } }"
		doc, err := ParseQueryWithLimits(&ast.Source{Name: "q", Input: input}, 0, 4)
		require.NoError(t, err)
		require.NotNil(t, doc)
	})
}

// FuzzParseQueryDepth exercises the parser against arbitrary input and asserts
// it always returns rather than crashing the process. With the depth guard in
// place, pathological nesting yields an error instead of a fatal stack overflow.
func FuzzParseQueryDepth(f *testing.F) {
	f.Add("{ a { b { c } } }")
	f.Add("{" + strings.Repeat("a{", 5000) + "x" + strings.Repeat("}", 5001))
	f.Add("query { ...frag } fragment frag on T { a { b } }")
	f.Add("{{{{{{{{{{")
	f.Add("")

	f.Fuzz(func(t *testing.T, input string) {
		// Must not panic or fatally overflow the stack for any input.
		_, _ = ParseQuery(&ast.Source{Name: "fuzz", Input: input})
	})
}
