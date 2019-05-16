package parser

import (
	"github.com/vektah/gqlparser/gqlerror"
	"testing"

	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/parser/testrunner"
)

func TestQueryDocument(t *testing.T) {
	testrunner.Test(t, "query_test.yml", func(t *testing.T, input string) testrunner.Spec {
		doc, err := ParseQuery(&ast.Source{Input: input, Name: "spec"})

		e, _ := err.(*gqlerror.Error)

		return testrunner.Spec{
			Error: e,
			AST:   ast.Dump(doc),
		}
	})
}
