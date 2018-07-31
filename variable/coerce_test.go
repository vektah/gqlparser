package variable_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/variable"
)

func TestCoerceValue(t *testing.T) {
	schema := gqlparser.MustLoadSchema(&ast.Source{
		Name:  "vars.graphql",
		Input: mustReadFile("./testdata/vars.graphql"),
	})

	t.Run("int", func(t *testing.T) {
		val := getValFromQuery(schema, `{ intArg(i: 1) }`)
		res, err := coerce(val)
		require.Nil(t, err)
		require.EqualValues(t, 1, res)
	})

	t.Run("string", func(t *testing.T) {
		val := getValFromQuery(schema, `{ stringArg(i: "hello") }`)
		res, err := coerce(val)
		require.Nil(t, err)
		require.EqualValues(t, "hello", res)
	})

	t.Run("null string", func(t *testing.T) {
		val := getValFromQuery(schema, `{ stringArg(i: null) }`)
		res, err := coerce(val)
		require.Nil(t, err)
		require.EqualValues(t, nil, res)
	})

	t.Run("float", func(t *testing.T) {
		val := getValFromQuery(schema, `{ floatArg(i: 1.2) }`)
		res, err := coerce(val)
		require.Nil(t, err)
		require.EqualValues(t, 1.2, res)
	})

	t.Run("int in float", func(t *testing.T) {
		val := getValFromQuery(schema, `{ floatArg(i: 1) }`)
		res, err := coerce(val)
		require.Nil(t, err)
		require.EqualValues(t, 1, res)
	})

	t.Run("bool", func(t *testing.T) {
		val := getValFromQuery(schema, `{ boolArg(i: true) }`)
		res, err := coerce(val)
		require.Nil(t, err)
		require.EqualValues(t, true, res)
	})

	t.Run("id int", func(t *testing.T) {
		val := getValFromQuery(schema, `{ idArg(i: 1234) }`)
		res, err := coerce(val)
		require.Nil(t, err)
		require.EqualValues(t, "1234", res)
	})

	t.Run("id string", func(t *testing.T) {
		val := getValFromQuery(schema, `{ idArg(i: "henlo") }`)
		res, err := coerce(val)
		require.Nil(t, err)
		require.EqualValues(t, "henlo", res)
	})

	t.Run("custom scalar", func(t *testing.T) {
		val := getValFromQuery(schema, `{ scalarArg(i: "henlo") }`)
		res, err := coerce(val)
		require.Nil(t, err)
		require.EqualValues(t, "henlo", res)
	})

	t.Run("objects", func(t *testing.T) {
		val := getValFromQuery(schema, `{ structArg(i: {name:"foo"}) }`)
		res, err := coerce(val)
		require.Nil(t, err)
		require.EqualValues(t, map[string]interface{}{"name": "foo"}, res)
	})

	t.Run("arrays", func(t *testing.T) {
		val := getValFromQuery(schema, `{ arrayArg(i: [{name:"foo"}]) }`)
		res, err := coerce(val)
		require.Nil(t, err)
		require.EqualValues(t, []interface{}{map[string]interface{}{"name": "foo"}}, res)
	})

	t.Run("variables", func(t *testing.T) {
		q := `query($id: Int!) { intArg(i: $id) }`
		val := getValFromQuery(schema, q)
		res, err := coerceWithVars(val, schema, q, map[string]interface{}{"id": 1})
		require.Nil(t, err)
		require.EqualValues(t, 1, res)
	})
}

func getValFromQuery(schema *ast.Schema, queryStr string) *ast.Value {
	q := gqlparser.MustLoadQuery(schema, queryStr, variable.DefaultInputCoercion)
	return q.Operations.ForName("").SelectionSet[0].(*ast.Field).Arguments[0].Value
}

func coerce(v *ast.Value) (interface{}, error) {
	return variable.NewEmptyBag(variable.DefaultInputCoercion).CoerceValue(v)
}

func coerceWithVars(v *ast.Value, schema *ast.Schema, queryStr string, vars map[string]interface{}) (interface{}, error) {
	q := gqlparser.MustLoadQuery(schema, queryStr, variable.DefaultInputCoercion)
	bag, err := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, vars)
	if err != nil {
		return nil, err
	}
	return bag.CoerceValue(v)
}
