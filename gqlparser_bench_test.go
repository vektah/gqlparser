package gqlparser_test

import (
	"testing"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

const (
	testSchema1 = `
type Color {
  hex: String!
  r: Int!
  g: Int!
  b: Int!
}

type Query {
  colors: [Color]
}`

	testQuery1 = `
query {
  colors {
    hex
    r
    g
    b
  }
}`
)

func BenchmarkLoadSchema(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gqlparser.LoadSchema(&ast.Source{Input: testSchema1})
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkLoadQuery(b *testing.B) {
	schema, err := gqlparser.LoadSchema(&ast.Source{Input: testSchema1})
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, errs := gqlparser.LoadQuery(schema, testQuery1)
		if errs != nil {
			b.Fatalf("unexpected errors: %v", errs)
		}
	}
}
