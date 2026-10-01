package gqlparser_test

import (
	"os"
	"testing"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator/rules"
)

// BenchmarkLoadSchema measures the cost a server pays once at startup: parse the SDL, then
// build and validate the schema.
func BenchmarkLoadSchema(b *testing.B) {
	input, err := os.ReadFile("validator/testdata/swapi.graphql")
	if err != nil {
		b.Fatal(err)
	}
	source := &ast.Source{Name: "swapi.graphql", Input: string(input)}
	b.ReportAllocs()

	for b.Loop() {
		if _, err := gqlparser.LoadSchema(source); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLoadQueryWithRules measures the library's front door: the function a server calls
// per request, parsing and validating in one step. Everything the parser and validator
// benchmarks measure separately is in here together, in the proportions a caller actually
// sees.
//
// LoadQueryWithRules rather than LoadQuery, which is deprecated. Benchmarking the deprecated
// spelling would pin the performance of the path new callers are told not to take.
func BenchmarkLoadQueryWithRules(b *testing.B) {
	input, err := os.ReadFile("validator/testdata/swapi.graphql")
	if err != nil {
		b.Fatal(err)
	}
	schema := gqlparser.MustLoadSchema(&ast.Source{Name: "swapi.graphql", Input: string(input)})
	defaultRules := rules.NewDefaultRules()

	for _, name := range benchmarkDocuments {
		b.Run(name, func(b *testing.B) {
			document := readDocument(b, "testdata", name)
			if _, errs := gqlparser.LoadQueryWithRules(
				schema,
				document,
				defaultRules,
			); len(
				errs,
			) > 0 {
				b.Fatalf("benchmark document %q does not validate: %s", name, errs)
			}
			b.ReportAllocs()

			for b.Loop() {
				gqlparser.LoadQueryWithRules(schema, document, defaultRules)
			}
		})
	}
}
