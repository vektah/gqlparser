package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vektah/gqlparser/v2/parser/testrunner"
)

func TestQueryDocument(t *testing.T) {
	testrunner.Test(t, "query_test.yml", func(t *testing.T, input string) testrunner.Spec {
		doc, err := ParseQuery(&ast.Source{Input: input, Name: "spec"})
		if err != nil {
			gqlErr := err.(*gqlerror.Error)
			return testrunner.Spec{
				Error: gqlErr,
				AST:   ast.Dump(doc),
			}
		}
		return testrunner.Spec{
			AST: ast.Dump(doc),
		}
	})
}

func TestQueryPosition(t *testing.T) {
	t.Run("query line number with comments", func(t *testing.T) {
		query, err := ParseQuery(&ast.Source{
			Input: `
	# comment 1
query SomeOperation {
	# comment 2
	myAction {
		id
	}
}
      `,
		})
		assert.NoError(t, err)
		assert.Equal(t, 3, query.Operations.ForName("SomeOperation").Position.Line)
		assert.Equal(t, 5, query.Operations.ForName("SomeOperation").SelectionSet[0].GetPosition().Line)
	})
}
