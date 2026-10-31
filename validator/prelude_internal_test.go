package validator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2/ast"
)

// PreludeDocument can't reach the panic, because the embedded prelude always
// parses, so this tests the helper it delegates to.
func TestMustParsePrelude(t *testing.T) {
	cases := map[string]struct {
		input     string
		wantPanic string
	}{
		"returns the parsed document": {
			input: "directive @d on FIELD",
		},
		"panics when the prelude does not parse": {
			input: "directive",
			wantPanic: "gqlparser: embedded prelude does not parse: " +
				"prelude.graphql:1:10: Expected @, found <EOF>",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			src := &ast.Source{Name: "prelude.graphql", Input: tc.input}
			if tc.wantPanic != "" {
				require.PanicsWithValue(t, tc.wantPanic, func() { mustParsePrelude(src) })
				return
			}
			doc := mustParsePrelude(src)
			require.NotNil(t, doc.Directives.ForName("d"))
		})
	}
}
