package proptest

import (
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
	"github.com/vektah/gqlparser/v2/validator/rules"
)

// TestValidationIsDeterministic checks that one source text validated twice, parsed afresh
// each time, reports the same errors in the same order.
//
// Several rules memoize across a document and iterate maps while doing it, and Go randomises
// map order on every run. A rule that leaks that order into its output produces a suite that
// passes locally and fails one run in ten on CI, and an error list that reorders itself
// between two requests for no reason the caller can see. The sequentialFieldsMap in
// OverlappingFieldsCanBeMerged exists because of this; nothing checked that the property holds
// for the rule set as a whole.
//
// Each pass parses again rather than reusing one document, because Walk annotates the AST it
// walks — it attaches each field's definition — and rules skip fields whose definition is not
// yet attached. Validating the same document twice therefore reports a superset the second
// time, which is a property of that mutation and not of the rules.
func TestValidationIsDeterministic(t *testing.T) {
	schema := gqlparser.MustLoadSchema(&ast.Source{
		Name: "property.graphqls",
		Input: `
			type Query { root: T, other: T }
			type T { x: T, y: T, a: String, b: String, fields: [T!] }
		`,
	})

	rapid.Check(t, func(t *rapid.T) {
		src := genValidatableQuery().Draw(t, "query")

		first := validate(schema, parseOrSkip(t, src))
		second := validate(schema, parseOrSkip(t, src))

		if first != second {
			t.Fatalf("validating\n%s\ntwice disagreed:\nfirst:\n%s\nsecond:\n%s",
				src, first, second)
		}
	})
}

func parseOrSkip(t *rapid.T, src string) *ast.QueryDocument {
	t.Helper()

	doc, err := parser.ParseQuery(&ast.Source{Name: "property.graphql", Input: src})
	if err != nil {
		t.Skipf("generated document does not parse: %s", err)
	}
	return doc
}

func validate(schema *ast.Schema, doc *ast.QueryDocument) string {
	var out strings.Builder
	for _, err := range validator.ValidateWithRules(schema, doc, rules.NewDefaultRules()) {
		out.WriteString(err.Error())
		out.WriteString("\n")
	}
	return out.String()
}

// genValidatableQuery draws documents against the schema above, aimed at the rules whose
// state spans a whole document: fragments reused and nested, fields that merge or conflict,
// and cycles. Documents that fail validation are as useful here as ones that pass, since the
// property is about the error list being stable rather than empty.
func genValidatableQuery() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		var doc strings.Builder

		roots := rapid.SliceOfN(rapid.SampledFrom([]string{"root", "other"}), 1, 2).
			Draw(t, "roots")
		doc.WriteString("{\n")
		for i, root := range roots {
			doc.WriteString("  r" + string(rune('A'+i)) + ": " + root + " " +
				genValidatableSelectionSet(2).Draw(t, "rootSelectionSet") + "\n")
		}
		doc.WriteString("}\n")

		for _, name := range rapid.SliceOfNDistinct(rapid.SampledFrom([]string{"F", "G", "H"}),
			0, 3, rapid.ID).Draw(t, "fragmentNames") {
			doc.WriteString("fragment " + name + " on T " +
				genValidatableSelectionSet(1).Draw(t, "fragmentSelectionSet") + "\n")
		}
		return doc.String()
	})
}

func genValidatableSelectionSet(depth int) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		selections := rapid.SliceOfN(genValidatableSelection(depth), 1, 3).Draw(t, "selections")
		return "{ " + strings.Join(selections, " ") + " }"
	})
}

func genValidatableSelection(depth int) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		switch rapid.SampledFrom([]string{"leaf", "nested", "spread", "inline"}).Draw(t, "kind") {
		case "spread":
			return "..." + rapid.SampledFrom([]string{"F", "G", "H"}).Draw(t, "spreadName")
		case "inline":
			if depth == 0 {
				return "a"
			}
			return "... on T " + genValidatableSelectionSet(depth-1).Draw(t, "inlineSelectionSet")
		case "nested":
			if depth == 0 {
				return "a"
			}
			// An alias shared between differing fields is what
			// OverlappingFieldsCanBeMerged reports, so draw it often.
			alias := rapid.SampledFrom([]string{"", "same: ", "same: "}).Draw(t, "alias")
			field := rapid.SampledFrom([]string{"x", "y", "fields"}).Draw(t, "nestedField")
			return alias + field + " " + genValidatableSelectionSet(depth-1).Draw(t, "subset")
		default:
			alias := rapid.SampledFrom([]string{"", "same: ", "same: "}).Draw(t, "alias")
			return alias + rapid.SampledFrom([]string{"a", "b"}).Draw(t, "leafField")
		}
	})
}
