package parser

import (
	"testing"

	"github.com/vektah/graphql-parser/spec"
)

func TestQueryDocument(t *testing.T) {
	spec.Test(t, "queryspec.yml", func(t *testing.T, input string) spec.Spec {
		doc, err := ParseQuery(input)
		return spec.Spec{
			Error: err,
			AST:   Dump(doc),
		}
	})
}
