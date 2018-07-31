package coerce_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/coerce"
)

func TestCoerceValue(t *testing.T) {
	schema := gqlparser.MustLoadSchema(&ast.Source{
		Name:  "vars.graphql",
		Input: mustReadFile("./testdata/vars.graphql"),
	})

	t.Run("int", func(t *testing.T) {
		field := getFieldFromQuery(schema, `{ intArg(i: 1) }`)
		res, err := coerceArgs(field, nil)
		require.Nil(t, err)
		require.EqualValues(t, 1, res["i"])
	})

	t.Run("string", func(t *testing.T) {
		field := getFieldFromQuery(schema, `{ stringArg(i: "hello") }`)
		res, err := coerceArgs(field, nil)
		require.Nil(t, err)
		require.EqualValues(t, "hello", res["i"])
	})

	t.Run("null string", func(t *testing.T) {
		field := getFieldFromQuery(schema, `{ stringArg(i: null) }`)
		res, err := coerceArgs(field, nil)
		require.Nil(t, err)
		require.EqualValues(t, nil, res["i"])
	})

	t.Run("float", func(t *testing.T) {
		field := getFieldFromQuery(schema, `{ floatArg(i: 1.2) }`)
		res, err := coerceArgs(field, nil)
		require.Nil(t, err)
		require.EqualValues(t, 1.2, res["i"])
	})

	t.Run("int in float", func(t *testing.T) {
		field := getFieldFromQuery(schema, `{ floatArg(i: 1) }`)
		res, err := coerceArgs(field, nil)
		require.Nil(t, err)
		require.EqualValues(t, 1, res["i"])
	})

	t.Run("bool", func(t *testing.T) {
		field := getFieldFromQuery(schema, `{ boolArg(i: true) }`)
		res, err := coerceArgs(field, nil)
		require.Nil(t, err)
		require.EqualValues(t, true, res["i"])
	})

	t.Run("id int", func(t *testing.T) {
		field := getFieldFromQuery(schema, `{ idArg(i: 1234) }`)
		res, err := coerceArgs(field, nil)
		require.Nil(t, err)
		require.EqualValues(t, "1234", res["i"])
	})

	t.Run("id string", func(t *testing.T) {
		field := getFieldFromQuery(schema, `{ idArg(i: "henlo") }`)
		res, err := coerceArgs(field, nil)
		require.Nil(t, err)
		require.EqualValues(t, "henlo", res["i"])
	})

	t.Run("custom scalar", func(t *testing.T) {
		field := getFieldFromQuery(schema, `{ scalarArg(i: "henlo") }`)
		res, err := coerceArgs(field, nil)
		require.Nil(t, err)
		require.EqualValues(t, "henlo", res["i"])
	})

	t.Run("objects", func(t *testing.T) {
		field := getFieldFromQuery(schema, `{ structArg(i: {name:"foo"}) }`)
		res, err := coerceArgs(field, nil)
		require.Nil(t, err)
		require.EqualValues(t, map[string]interface{}{"name": "foo"}, res["i"])
	})

	t.Run("arrays", func(t *testing.T) {
		field := getFieldFromQuery(schema, `{ arrayArg(i: [{name:"foo"}]) }`)
		res, err := coerceArgs(field, nil)
		require.Nil(t, err)
		require.EqualValues(t, []interface{}{map[string]interface{}{"name": "foo"}}, res["i"])
	})

	t.Run("variables", func(t *testing.T) {
		q := `query($id: Int!) { intArg(i: $id) }`
		field := getFieldFromQuery(schema, q)
		res, err := coerceArgs(field, map[string]interface{}{"id": 1})
		require.Nil(t, err)
		require.EqualValues(t, 1, res["i"])
	})
}

func getFieldFromQuery(schema *ast.Schema, queryStr string) *ast.Field {
	q := gqlparser.MustLoadQuery(schema, queryStr, coerce.DefaultScalar)
	return q.Operations.ForName("").SelectionSet[0].(*ast.Field)
}

func coerceArgs(field *ast.Field, vars map[string]interface{}) (map[string]interface{}, error) {
	return coerce.FieldArguments(field, vars, coerce.DefaultScalar)
}
