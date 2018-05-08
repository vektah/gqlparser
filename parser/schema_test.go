package parser

import (
	"testing"

	"github.com/vektah/graphql-parser/spec"
)

func TestSchemaDocument(t *testing.T) {
	spec.Test(t, "../spec/schema.yml", func(t *testing.T, input string) spec.Spec {
		doc, err := ParseSchema(input)
		return spec.Spec{
			Error: err,
			AST:   Dump(doc),
		}
	})
}
