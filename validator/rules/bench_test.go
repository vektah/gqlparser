package rules_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
	"github.com/vektah/gqlparser/v2/validator/core"
	"github.com/vektah/gqlparser/v2/validator/rules"
)

// BenchmarkRule measures one rule at a time against the introspection query.
//
// Validating with the whole rule set says what a request costs; it does not say which rule to
// look at when that cost moves. These are the rules whose cost is not proportional to the
// document: they carry state across it, memoize, or compare every field against every other.
// MaxIntrospectionDepth and OverlappingFieldsCanBeMerged have both been exponential in a
// released version, and a benchmark per rule is what turns the next such change from a
// mystery in the aggregate number into a named line.
//
// The introspection query is the input because it is both the largest document most servers
// see and the one these rules do the most work on.
func BenchmarkRule(b *testing.B) {
	schema := benchSchema(b)
	document := readDocument(b, "introspection")

	for _, rule := range []struct {
		name string
		rule core.Rule
	}{
		{"MaxIntrospectionDepth", rules.MaxIntrospectionDepth},
		{"OverlappingFieldsCanBeMerged", rules.OverlappingFieldsCanBeMergedRule},
		{"ValuesOfCorrectType", rules.ValuesOfCorrectTypeRule},
		{"FieldsOnCorrectType", rules.FieldsOnCorrectTypeRule},
		{"KnownArgumentNames", rules.KnownArgumentNamesRule},
		{"NoFragmentCycles", rules.NoFragmentCyclesRule},
	} {
		b.Run(rule.name, func(b *testing.B) {
			ruleSet := rules.NewRules(rule.rule)
			src := &ast.Source{Name: "bench.graphql", Input: document}
			b.ReportAllocs()

			for b.Loop() {
				// Reparsed per iteration: Walk annotates the document, so a reused one is
				// not the input a request presents.
				b.StopTimer()
				doc, err := parser.ParseQuery(src)
				if err != nil {
					b.Fatal(err)
				}
				b.StartTimer()

				validator.ValidateWithRules(schema, doc, ruleSet)
			}
		})
	}
}

// BenchmarkFragmentHeavyRules measures the rules that carry state across a document against
// one that reuses fragments, which is the shape that has historically made them superlinear.
func BenchmarkFragmentHeavyRules(b *testing.B) {
	schema := benchSchema(b)
	src := &ast.Source{Name: "bench.graphql", Input: readDocument(b, "fragments")}
	ruleSet := rules.NewRules(
		rules.OverlappingFieldsCanBeMergedRule,
		rules.NoFragmentCyclesRule,
		rules.NoUnusedFragmentsRule,
		rules.KnownFragmentNamesRule,
	)
	b.ReportAllocs()

	for b.Loop() {
		b.StopTimer()
		doc, err := parser.ParseQuery(src)
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		validator.ValidateWithRules(schema, doc, ruleSet)
	}
}

func benchSchema(tb testing.TB) *ast.Schema {
	tb.Helper()

	input, err := os.ReadFile(filepath.Join("..", "testdata", "swapi.graphql"))
	if err != nil {
		tb.Fatalf("reading the benchmark schema: %s", err)
	}
	return gqlparser.MustLoadSchema(&ast.Source{Name: "swapi.graphql", Input: string(input)})
}

func readDocument(tb testing.TB, name string) string {
	tb.Helper()

	input, err := os.ReadFile(filepath.Join("..", "..", "testdata", name+".graphql"))
	if err != nil {
		tb.Fatalf("reading the benchmark document: %s", err)
	}
	return string(input)
}

// BenchmarkFieldMerging measures comparing fields that share a response name.
//
// OverlappingFieldsCanBeMerged only reaches its argument comparison when two selections
// collide on a name, so a document of distinct fields never executes it however large it is.
// That comparison walks composite argument values element by element, and it is quadratic in
// the number of colliding selections, which is why it is worth measuring on its own rather
// than inside a document where it never runs.
func BenchmarkFieldMerging(b *testing.B) {
	schema := gqlparser.MustLoadSchema(&ast.Source{Name: "merge.graphqls", Input: `
		type Query { search(filter: Filter, codes: [String!]): Result }
		type Result { id: ID!, name: String }
		input Filter { name: String, nested: Filter }
	`})

	// Identical arguments, so every pair compares all the way down rather than failing at the
	// first difference — the expensive direction, and the one a valid document takes.
	const argument = `(filter: { name: "a", nested: { name: "b" } }, codes: ["x", "y"])`
	for _, selections := range []int{2, 8, 32} {
		var document strings.Builder
		document.WriteString("{\n")
		for range selections {
			document.WriteString("  same: search" + argument + " { id name }\n")
		}
		document.WriteString("}\n")

		src := &ast.Source{Name: "merge.graphql", Input: document.String()}
		ruleSet := rules.NewRules(rules.OverlappingFieldsCanBeMergedRule)

		b.Run("selections="+strconv.Itoa(selections), func(b *testing.B) {
			doc, err := parser.ParseQuery(src)
			if err != nil {
				b.Fatal(err)
			}
			if errs := validator.ValidateWithRules(schema, doc, ruleSet); len(errs) > 0 {
				b.Fatalf("the benchmark document should merge cleanly: %s", errs)
			}
			b.ReportAllocs()

			for b.Loop() {
				b.StopTimer()
				doc, err := parser.ParseQuery(src)
				if err != nil {
					b.Fatal(err)
				}
				b.StartTimer()

				validator.ValidateWithRules(schema, doc, ruleSet)
			}
		})
	}
}
