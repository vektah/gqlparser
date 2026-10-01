package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

func BenchmarkParseQuery(b *testing.B) {
	for _, name := range benchmarkDocuments {
		b.Run(name, func(b *testing.B) {
			src := &ast.Source{Name: "bench.graphql", Input: readDocument(b, name)}
			b.ReportAllocs()

			for b.Loop() {
				if _, err := parser.ParseQuery(src); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkParseSchema reads a schema of the size a real service has. A server parses one at
// startup rather than per request, so this is measured apart from the query path.
func BenchmarkParseSchema(b *testing.B) {
	input, err := os.ReadFile("../validator/testdata/swapi.graphql")
	if err != nil {
		b.Fatal(err)
	}
	src := &ast.Source{Name: "swapi.graphql", Input: string(input)}
	b.ReportAllocs()

	for b.Loop() {
		if _, err := parser.ParseSchema(src); err != nil {
			b.Fatal(err)
		}
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
