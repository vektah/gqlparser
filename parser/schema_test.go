package parser

import (
	"github.com/vektah/gqlparser/v2/gqlerror"
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser/testrunner"
)

func TestSchemaDocument(t *testing.T) {
	testrunner.Test(t, "schema_test.yml", func(t *testing.T, input string) testrunner.Spec {
		doc, err := ParseSchema(&ast.Source{Input: input, Name: "spec"})
		if err != nil {
			return testrunner.Spec{
				Error: err.(*gqlerror.Error),
				AST:   ast.Dump(doc),
			}
		}
		return testrunner.Spec{
			AST:   ast.Dump(doc),
		}
	})
}
