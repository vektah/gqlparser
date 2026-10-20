package proptest

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"unicode/utf8"

	"pgregory.net/rapid"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/lexer"
)

// TestBracedUnicodeEscapeDecodes checks that a braced unicode escape lexes to the code point
// its digits spell when that is a Unicode scalar value and there are at most eight digits, as
// graphql-js allows, and is an error otherwise.
//
// The lexer accumulates the digits in a rune, which is 32 bits wide, so eight digits can set
// its sign bit and nine can shift a high digit out altogether: \u{100000041} would come out as
// "A" if the cap did not stop it. The oracle here never accumulates anything; it starts from
// the code point and prints the digits, so it cannot share that mistake. Code points are drawn
// from three ranges because rapid favours small values: the scalar values, which a uniform
// 32-bit value almost never hits, any value at all, and the values with the sign bit set,
// which rapid otherwise almost never reaches.
func TestBracedUnicodeEscapeDecodes(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		x := rapid.OneOf(
			rapid.Uint32Range(0, utf8.MaxRune),
			rapid.Uint32(),
			rapid.Uint32Range(1<<31, math.MaxUint32),
		).Draw(t, "codepoint")
		width := rapid.IntRange(0, 12).Draw(t, "width")
		lower := rapid.Bool().Draw(t, "lower")

		digits := fmt.Sprintf("%0*X", width, x)
		if lower {
			digits = strings.ToLower(digits)
		}
		input := `"\u{` + digits + `}"`

		l := lexer.New(&ast.Source{Name: "property.graphql", Input: input})
		tok, err := l.ReadToken()

		r := rune(x)
		if len(digits) > 8 || !utf8.ValidRune(r) {
			if err == nil {
				t.Fatalf("%s lexed as %q; wanted an error", input, tok.Value)
			}
			return
		}
		if err != nil {
			t.Fatalf("%s does not lex: %s", input, err)
		}
		if want := string(r); tok.Value != want {
			t.Fatalf("%s lexed as %q; wanted %q", input, tok.Value, want)
		}
	})
}
