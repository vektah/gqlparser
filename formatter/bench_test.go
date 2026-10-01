package formatter_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/formatter"
	"github.com/vektah/gqlparser/v2/parser"
)

// BenchmarkFormatQueryDocument parses once outside the loop: formatting does not modify the
// document, so one parse serves every iteration.
func BenchmarkFormatQueryDocument(b *testing.B) {
	for _, name := range benchmarkDocuments {
		b.Run(name, func(b *testing.B) {
			doc, err := parser.ParseQuery(
				&ast.Source{Name: "bench.graphql", Input: readDocument(b, name)},
			)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()

			for b.Loop() {
				formatter.NewFormatter(io.Discard).FormatQueryDocument(doc)
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
