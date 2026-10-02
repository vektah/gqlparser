package proptest

import (
	"testing"

	"pgregory.net/rapid"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

// TestStringValueRoundTrips checks that printing a string value and parsing the result
// recovers the value that went in.
//
// This is the guarantee that makes the formatter safe to run over a document somebody else
// wrote: a printer that cannot read its own output silently edits the query. It used to fail,
// because String() called strconv.Quote, which spells a control character \x1b and an
// unprintable astral one \U000e0001 — Go escapes the GraphQL grammar has no rule for. The
// golden files cover the strings someone thought to write down; this covers the ones nobody did.
//
// A counterexample is either a string the formatter prints in a form the lexer rejects, or
// one it prints in a form the lexer reads back as something else. The second is the dangerous
// kind, and the kind a golden file is least likely to catch.
func TestStringValueRoundTrips(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		raw := rapid.String().Draw(t, "raw")

		quoted := (&ast.Value{Kind: ast.StringValue, Raw: raw}).String()

		doc, err := parser.ParseQuery(&ast.Source{
			Name:  "property.graphql",
			Input: "{ f(a: " + quoted + ") }",
		})
		if err != nil {
			t.Fatalf("%q printed as %s, which does not parse: %s", raw, quoted, err)
		}

		got := doc.Operations[0].SelectionSet[0].(*ast.Field).Arguments[0].Value.Raw
		if got != raw {
			t.Fatalf("%q printed as %s, which read back as %q", raw, quoted, got)
		}
	})
}

// TestStringValueQuotingIsIndependentOfContext checks that a character is printed the same way
// whether or not something else in the same string needs escaping.
//
// quoteString escapes nothing until it meets a byte that has to be escaped, so a string takes
// one of two routes through it. That is an optimisation, and an optimisation that changes the
// answer is a bug: a character's printed form must not depend on its neighbours. Prefixing a
// NUL forces the escaping route without altering what follows it, so the two routes can be
// compared directly on the same input.
//
// Drawing bytes rather than runes is the point of this property. ast.Value.Raw is a Go string
// and nothing stops a caller — gqlgen builds these by hand — from putting bytes in it that are
// not valid UTF-8. Round-tripping cannot reach that input: the lexer replaces malformed bytes
// when it decodes escapes, so TestStringValueRoundTrips can only speak for valid UTF-8. This
// property holds for any bytes at all, and it is the one that fails when the two routes are
// written against different alphabets — one walking bytes, the other runes.
func TestStringValueQuotingIsIndependentOfContext(t *testing.T) {
	// A NUL is printed as \u0000 and is a complete one-byte sequence, so it cannot combine
	// with whatever follows it the way a lead byte such as 0xC2 would.
	const nul = "\x00"

	quote := func(raw string) string {
		return (&ast.Value{Kind: ast.StringValue, Raw: raw}).String()
	}

	rapid.Check(t, func(t *rapid.T) {
		raw := string(rapid.SliceOfN(rapid.Byte(), 0, 64).Draw(t, "bytes"))

		alone := quote(raw)
		// alone[1:] is the printed body plus its closing quote.
		want := `"\u0000` + alone[1:]

		if got := quote(nul + raw); got != want {
			t.Fatalf("%q printed as %s alone but as %s after a NUL; wanted %s",
				raw, alone, got, want)
		}
	})
}
