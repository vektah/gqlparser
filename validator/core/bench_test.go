package core_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator/core"
)

// BenchmarkWalk measures the traversal every rule rides on. Each rule registers observers and
// Walk calls them once per node, so this is the floor under the whole validator: no rule can
// be cheaper than the walk that reaches it.
//
// The observer here does nothing, which is the point — what is being measured is the
// traversal and the type resolution it does on the way, not any rule's opinion.
func BenchmarkWalk(b *testing.B) {
	schema := benchSchema(b)

	for _, name := range benchmarkDocuments {
		b.Run(name, func(b *testing.B) {
			src := &ast.Source{Name: "bench.graphql", Input: readDocument(b, name)}
			b.ReportAllocs()

			for b.Loop() {
				// Parsed inside the loop because Walk annotates the document it walks, so a
				// reused one takes a path no caller takes.
				b.StopTimer()
				doc, err := parser.ParseQuery(src)
				if err != nil {
					b.Fatal(err)
				}
				observers := &core.Events{}
				observers.OnField(func(*core.Walker, *ast.Field) {})
				b.StartTimer()

				core.Walk(schema, doc, observers)
			}
		})
	}
}

// BenchmarkSuggestionList measures the "did you mean" search, which runs a Levenshtein
// distance against every candidate name. It is on the error path only, but a schema with a
// thousand types pays it per unknown field, and a client that sends many bad requests is the
// one most likely to be hostile.
func BenchmarkSuggestionList(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		options := make([]string, size)
		for i := range options {
			options[i] = "someFieldName" + string(rune('a'+i%26)) + string(rune('a'+i/26%26))
		}

		b.Run("options="+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				core.SuggestionList("someFieldNamez", options)
			}
		})
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

// benchmarkDocuments are the shapes measured against; they live in the repository's testdata
// so nothing a benchmark needs is compiled into the library.
var benchmarkDocuments = []string{"small", "fragments", "deep", "introspection"}

func readDocument(tb testing.TB, name string) string {
	tb.Helper()

	input, err := os.ReadFile(filepath.Join("..", "..", "testdata", name+".graphql"))
	if err != nil {
		tb.Fatalf("reading the benchmark document: %s", err)
	}
	return string(input)
}
