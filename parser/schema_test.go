package parser

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vektah/gqlparser/v2/parser/testrunner"
)

func TestSchemaDocument(t *testing.T) {
	testrunner.Test(t, "schema_test.yml", func(t *testing.T, input string) testrunner.Spec {
		doc, err := ParseSchema(&ast.Source{Input: input, Name: "spec"})
		if err != nil {
			return testrunner.Spec{
				Error: func() *gqlerror.Error {
					target := &gqlerror.Error{}
					_ = errors.As(err, &target)
					return target
				}(),
				AST: ast.Dump(doc),
			}
		}
		return testrunner.Spec{
			AST: ast.Dump(doc),
		}
	})
}

func TestFieldPositionWithBlockStringDescription(t *testing.T) {
	schema, parseErr := ParseSchema(&ast.Source{
		Input: `type Query {
					"""
					multi
					line
					"""
					myField: String
				}`,
	})
	assert.NoError(t, parseErr)
	field := schema.Definitions.ForName("Query").Fields.ForName("myField")
	assert.Equal(t, 6, field.Position.Line)
	assert.Equal(t, 6, field.Position.Column)
}

func TestArgumentPositionWithBlockStringDescription(t *testing.T) {
	schema, parseErr := ParseSchema(&ast.Source{
		Input: `type Query {
					myField(
						"""
						multi
						line
						"""
						myArg: String
					): String
				}`,
	})
	assert.NoError(t, parseErr)
	arg := schema.Definitions.ForName("Query").Fields.ForName("myField").Arguments.ForName("myArg")
	assert.Equal(t, 7, arg.Position.Line)
	assert.Greater(t, arg.Position.Column, 0)
}

func TestInputValuePositionWithBlockStringDescription(t *testing.T) {
	schema, parseErr := ParseSchema(&ast.Source{
		Input: `input MyInput {
					"""
					multi
					line
					"""
					myField: String
				}`,
	})
	assert.NoError(t, parseErr)
	field := schema.Definitions.ForName("MyInput").Fields.ForName("myField")
	assert.Equal(t, 6, field.Position.Line)
	assert.Greater(t, field.Position.Column, 0)
}

func TestEnumValuePositionWithBlockStringDescription(t *testing.T) {
	schema, parseErr := ParseSchema(&ast.Source{
		Input: `enum MyEnum {
					"""
					multi
					line
					"""
					MY_VALUE
				}`,
	})
	assert.NoError(t, parseErr)
	val := schema.Definitions.ForName("MyEnum").EnumValues.ForName("MY_VALUE")
	assert.Equal(t, 6, val.Position.Line)
	assert.Greater(t, val.Position.Column, 0)
}

func TestTypePosition(t *testing.T) {
	t.Run("type line number with no bang", func(t *testing.T) {
		schema, parseErr := ParseSchema(&ast.Source{
			Input: `type query {
						me: User
					}
			`,
		})
		assert.NoError(t, parseErr)
		assert.Equal(
			t,
			2,
			schema.Definitions.ForName("query").Fields.ForName("me").Type.Position.Line,
		)
	})
	t.Run("type line number with bang", func(t *testing.T) {
		schema, parseErr := ParseSchema(&ast.Source{
			Input: `type query {
						me: User!
					}
			`,
		})
		assert.NoError(t, parseErr)
		assert.Equal(
			t,
			2,
			schema.Definitions.ForName("query").Fields.ForName("me").Type.Position.Line,
		)
	})
	t.Run("type line number with comments", func(t *testing.T) {
		schema, parseErr := ParseSchema(&ast.Source{
			Input: `type query {
            # comment
						me: User
					}
			`,
		})
		assert.NoError(t, parseErr)
		assert.Equal(
			t,
			3,
			schema.Definitions.ForName("query").Fields.ForName("me").Type.Position.Line,
		)
	})
}
