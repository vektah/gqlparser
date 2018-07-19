package parser

import (
	"testing"

	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/spec"
)

func TestSchemaDocument(t *testing.T) {
	spec.Test(t, "../spec/schema.yml", func(t *testing.T, input string) spec.Spec {
		doc, err := ParseSchema(&ast.Source{Input: input, Name: "spec"})
		return spec.Spec{
			Error: err,
			AST:   spec.DumpAST(doc),
		}
	})
}
