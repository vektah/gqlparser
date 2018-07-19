package parser

import (
	"testing"

	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/spec"
)

func TestQueryDocument(t *testing.T) {
	spec.Test(t, "../spec/query.yml", func(t *testing.T, input string) spec.Spec {
		doc, err := ParseQuery(&ast.Source{Input: input, Name: "spec"})
		return spec.Spec{
			Error: err,
			AST:   spec.DumpAST(doc),
		}
	})
}
