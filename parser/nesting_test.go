package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func TestMaxNestingDepth(t *testing.T) {
	// Each build func returns a document that nests exactly depth levels,
	// counting the operation's own selection set where there is one.
	cases := map[string]struct {
		schema bool
		build  func(depth int) string
	}{
		"selection set": {build: func(d int) string {
			return "{" + strings.Repeat("a{", d-1) + "x" + strings.Repeat("}", d)
		}},
		"inline fragment": {build: func(d int) string {
			return "{" + strings.Repeat("... on T {", d-1) + "x" + strings.Repeat("}", d)
		}},
		"fragment definition": {build: func(d int) string {
			return "fragment F on T {" + strings.Repeat("a{", d-1) + "x" + strings.Repeat("}", d)
		}},
		"list argument": {build: func(d int) string {
			return "{ f(a: " + strings.Repeat("[", d-1) + strings.Repeat("]", d-1) + ") }"
		}},
		"object argument": {build: func(d int) string {
			return "{ f(a: " + strings.Repeat("{a:", d-1) + "1" + strings.Repeat("}", d-1) + ") }"
		}},
		"directive argument": {build: func(d int) string {
			return "{ x @d(a: " + strings.Repeat("[", d-1) + strings.Repeat("]", d-1) + ") }"
		}},
		"variable type": {build: func(d int) string {
			return "query($v: " + strings.Repeat("[", d) + "T" + strings.Repeat("]", d) + ") { x }"
		}},
		"variable default": {build: func(d int) string {
			return "query($v: T = " + strings.Repeat("[", d) + strings.Repeat("]", d) + ") { x }"
		}},
		"selection set and list combined": {build: func(d int) string {
			half := d / 2
			return "{" + strings.Repeat("a{", half-1) +
				"f(a: " + strings.Repeat("[", d-half) + strings.Repeat("]", d-half) + ")" +
				strings.Repeat("}", half)
		}},
		"schema field type": {schema: true, build: func(d int) string {
			return "type Q { f: " + strings.Repeat("[", d) + "T" + strings.Repeat("]", d) + " }"
		}},
		"schema argument default": {schema: true, build: func(d int) string {
			return "type Q { f(a: T = " + strings.Repeat("[", d) + strings.Repeat("]", d) + "): T }"
		}},
	}
	parse := func(schema bool, input string) error {
		src := &ast.Source{Name: "q", Input: input}
		if schema {
			_, err := ParseSchema(src)
			return err
		}
		_, err := ParseQuery(src)
		return err
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, parse(tc.schema, tc.build(MaxNestingDepth)))

			// One level over the limit is rejected with a located error.
			err := parse(tc.schema, tc.build(MaxNestingDepth+1))
			var gqlErr *gqlerror.Error
			require.ErrorAs(t, err, &gqlErr)
			require.Equal(t, "exceeded maximum nesting depth of 1000", gqlErr.Message)
			require.NotEmpty(t, gqlErr.Locations)

			// Deep enough to overflow the goroutine stack without the limit.
			err = parse(tc.schema, tc.build(3_000_000))
			require.ErrorContains(t, err, "exceeded maximum nesting depth")
		})
	}
}

func TestMaxNestingDepthCountsCurrentPathOnly(t *testing.T) {
	// Sibling branches each 600 deep: fine unless depth leaks between them.
	branch := strings.Repeat("a{", 599) + "x" + strings.Repeat("}", 599)
	list := strings.Repeat("[", 600) + strings.Repeat("]", 600)
	cases := map[string]string{
		"sibling fields":     "{" + branch + " " + branch + "}",
		"sibling operations": "query A {" + branch + "} query B {" + branch + "}",
		"sibling arguments":  "{ f(a: " + list + ", b: " + list + ") }",
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ParseQuery(&ast.Source{Input: input})
			require.NoError(t, err)
		})
	}
}
