package gqlparser_test

import (
	"os"
	"path/filepath"
	"testing"
)

// benchmarkDocuments are the shapes the benchmarks measure against: the one most requests
// have, one that leans on fragments, one that nests, and the introspection query every client
// sends on connect.
//
// They live in testdata rather than in a package of constants so that nothing a benchmark
// needs is compiled into the library consumers import. Each package that benchmarks reads
// them with its own copy of this loader, which is six lines and keeps the benchmark readable
// without a shared helper package to jump to.
var benchmarkDocuments = []string{"small", "fragments", "deep", "introspection"}

func readDocument(tb testing.TB, dir, name string) string {
	tb.Helper()

	input, err := os.ReadFile(filepath.Join(dir, name+".graphql"))
	if err != nil {
		tb.Fatalf("reading the benchmark document: %s", err)
	}
	return string(input)
}
