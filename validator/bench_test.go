package validator_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
	"github.com/vektah/gqlparser/v2/validator/rules"
)

// BenchmarkParseAndValidateQuery measures what a server does per request: parse the document,
// then validate it against an already-loaded schema.
//
// Parsing is inside the loop because Walk annotates the document it walks, attaching each
// field's definition. Validating one document repeatedly would measure the annotated path from
// the second iteration on, which no request ever takes. The parser benchmarks measure parsing
// on its own, so the difference between them is validation's share.
func BenchmarkParseAndValidateQuery(b *testing.B) {
	input, err := os.ReadFile("testdata/swapi.graphql")
	if err != nil {
		b.Fatal(err)
	}
	schema := gqlparser.MustLoadSchema(&ast.Source{Name: "swapi.graphql", Input: string(input)})
	defaultRules := rules.NewDefaultRules()

	for _, name := range benchmarkDocuments {
		b.Run(name, func(b *testing.B) {
			src := &ast.Source{Name: "bench.graphql", Input: readDocument(b, name)}

			// A document that fails validation measures the error path and the formatting
			// of messages, which is not what this is for.
			doc, err := parser.ParseQuery(src)
			if err != nil {
				b.Fatal(err)
			}
			if errs := validator.ValidateWithRules(schema, doc, defaultRules); len(errs) > 0 {
				b.Fatalf("benchmark document %q does not validate: %s", name, errs)
			}
			b.ReportAllocs()

			for b.Loop() {
				doc, err := parser.ParseQuery(src)
				if err != nil {
					b.Fatal(err)
				}
				validator.ValidateWithRules(schema, doc, defaultRules)
			}
		})
	}
}

// benchmarkDocuments are the shapes measured against; they live in the repository's testdata
// so nothing a benchmark needs is compiled into the library.
var benchmarkDocuments = []string{"small", "fragments", "deep", "introspection"}

func readDocument(tb testing.TB, name string) string {
	tb.Helper()

	input, err := os.ReadFile(filepath.Join("..", "testdata", name+".graphql"))
	if err != nil {
		tb.Fatalf("reading the benchmark document: %s", err)
	}
	return string(input)
}

// BenchmarkVariableValues measures coercing a request's variables against the operation's
// declared types.
//
// Every request that carries variables pays this, after validation and before execution, and
// it is the one part of the per-request path that scales with the size of the input data
// rather than the size of the query: a list of a thousand input objects is coerced element by
// element. The sizes below are what makes that visible.
func BenchmarkVariableValues(b *testing.B) {
	schema := gqlparser.MustLoadSchema(&ast.Source{Name: "vars.graphqls", Input: `
		type Query { search(filter: Filter, ids: [ID!], first: Int): String }
		input Filter { name: String, episode: Episode, nested: Filter }
		enum Episode { NEWHOPE EMPIRE JEDI }
	`})
	// Loaded rather than parsed: VariableValues reads each variable's resolved definition,
	// which validation is what attaches. Handed a merely parsed document it dereferences a
	// nil definition, so this is the state a caller must reach it in.
	doc, errs := gqlparser.LoadQueryWithRules(schema, `
		query Search($filter: Filter, $ids: [ID!], $first: Int) {
			search(filter: $filter, ids: $ids, first: $first)
		}
	`, rules.NewDefaultRules())
	if len(errs) > 0 {
		b.Fatalf("the benchmark query does not validate: %s", errs)
	}
	operation := doc.Operations.ForName("Search")

	for _, size := range []int{1, 100, 1000} {
		ids := make([]any, size)
		for i := range ids {
			ids[i] = strconv.Itoa(i)
		}
		variables := map[string]any{
			"first": size,
			"ids":   ids,
			"filter": map[string]any{
				"name":    "luke",
				"episode": "JEDI",
				"nested":  map[string]any{"name": "leia"},
			},
		}

		b.Run("ids="+strconv.Itoa(size), func(b *testing.B) {
			if _, err := validator.VariableValues(schema, operation, variables); err != nil {
				b.Fatalf("the benchmark variables do not coerce: %s", err)
			}
			b.ReportAllocs()

			for b.Loop() {
				if _, err := validator.VariableValues(schema, operation, variables); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
