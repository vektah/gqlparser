package lexer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/lexer"
)

// BenchmarkReadToken measures tokenising alone, which every other path pays for before it
// does anything of its own.
func BenchmarkReadToken(b *testing.B) {
	for _, name := range benchmarkDocuments {
		b.Run(name, func(b *testing.B) {
			src := &ast.Source{Name: "bench.graphql", Input: readDocument(b, name)}
			b.ReportAllocs()

			for b.Loop() {
				lex := lexer.New(src)
				for {
					token, err := lex.ReadToken()
					if err != nil {
						b.Fatal(err)
					}
					if token.Kind == lexer.EOF {
						break
					}
				}
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
