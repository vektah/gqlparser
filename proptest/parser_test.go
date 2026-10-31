package proptest

import (
	"testing"

	"pgregory.net/rapid"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

// TestParseQueryAlwaysReturnsADocument checks that parsing arbitrary bytes never panics and
// always yields a document to return, however malformed the input.
//
// ParseQuery returns the document it managed to build alongside any error, so a caller may
// dereference the result on the error path — ast.QueryDocument's zero value is usable, and
// several callers walk it to report where parsing stopped. That makes a nil document a
// crash in someone else's process rather than an error in ours, which is worth pinning.
//
// The parser is also the first thing a server points at input it did not write, so the shape
// of a malformed document is chosen by whoever is attacking it rather than by whoever wrote
// the corpus. Fuzz.spec.yml pins the inputs that have broken it before; this reaches the ones
// nobody has written down.
func TestParseQueryAlwaysReturnsADocument(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		input := rapid.OneOf(
			rapid.String(),
			// Bytes drawn freely reach invalid UTF-8 and lone surrogates, which a generator
			// of Go strings alone rarely produces.
			rapid.Custom(func(t *rapid.T) string {
				return string(rapid.SliceOfN(rapid.Byte(), 0, 64).Draw(t, "bytes"))
			}),
			// Fragments of real syntax are far likelier to reach the parser's interesting
			// paths than random bytes, which it rejects at the first token.
			rapid.Custom(func(t *rapid.T) string {
				pieces := rapid.SliceOfN(rapid.SampledFrom([]string{
					"{", "}", "(", ")", "[", "]", "...", "on", "query", "fragment", "$v",
					":", ",", "@dir", `"s"`, `"""`, "1", "a", "...F", "!", "=", "|", "&",
				}), 0, 12).Draw(t, "pieces")
				out := ""
				for _, piece := range pieces {
					out += piece + " "
				}
				return out
			}),
		).Draw(t, "input")

		doc, err := parser.ParseQuery(&ast.Source{Name: "property.graphql", Input: input})

		if doc == nil {
			t.Fatalf("parsing %q returned no document; error was %v", input, err)
		}
	})
}
