package coerce_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/coerce"
)

func TestCoerceVars(t *testing.T) {
	schema := gqlparser.MustLoadSchema(&ast.Source{
		Name:  "vars.graphql",
		Input: mustReadFile("./testdata/vars.graphql"),
	})

	t.Run("undefined variable", func(t *testing.T) {
		t.Run("without default", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query($id: Int!) { intArg(i: $id) }`, coerce.DefaultScalar)
			_, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, nil)
			require.EqualError(t, gerr, "input: variable.id must be defined")
		})

		t.Run("nil in required value", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query($id: Int!) { intArg(i: $id) }`, coerce.DefaultScalar)
			_, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, map[string]interface{}{
				"id": nil,
			})
			require.EqualError(t, gerr, "input: variable.id cannot be null")
		})

		t.Run("with default", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query($id: Int! = 1) { intArg(i: $id) }`, coerce.DefaultScalar)
			vars, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, nil)
			require.Nil(t, gerr)
			require.EqualValues(t, 1, vars["id"])
		})

		t.Run("with union", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query($id: Int! = 1) { intArg(i: $id) }`, coerce.DefaultScalar)
			vars, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, nil)
			require.Nil(t, gerr)
			require.EqualValues(t, 1, vars["id"])
		})
	})

	t.Run("input object", func(t *testing.T) {
		t.Run("non object", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`, coerce.DefaultScalar)
			_, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, map[string]interface{}{
				"var": "hello",
			})
			require.EqualError(t, gerr, "input: variable.var must be a InputType")
		})

		t.Run("defaults", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType! = {name: "foo"}) { structArg(i: $var) }`, coerce.DefaultScalar)
			vars, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, nil)
			require.Nil(t, gerr)
			require.EqualValues(t, map[string]interface{}{"name": "foo"}, vars["var"])
		})

		t.Run("valid value", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`, coerce.DefaultScalar)
			vars, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, map[string]interface{}{
				"var": map[string]interface{}{
					"name": "foobar",
				},
			})
			require.Nil(t, gerr)
			require.EqualValues(t, map[string]interface{}{"name": "foobar"}, vars["var"])
		})

		t.Run("missing required values", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`, coerce.DefaultScalar)
			_, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, map[string]interface{}{
				"var": map[string]interface{}{},
			})
			require.EqualError(t, gerr, "input: variable.var.name must be defined")
		})

		t.Run("null required field", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`, coerce.DefaultScalar)
			_, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, map[string]interface{}{
				"var": map[string]interface{}{
					"name": nil,
				},
			})
			require.EqualError(t, gerr, "input: variable.var.name cannot be null")
		})

		t.Run("unknown field", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`, coerce.DefaultScalar)
			_, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, map[string]interface{}{
				"var": map[string]interface{}{
					"name":    "foobar",
					"foobard": true,
				},
			})
			require.EqualError(t, gerr, "input: variable.var.foobard unknown field")
		})
	})

	t.Run("array", func(t *testing.T) {
		t.Run("non array", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: [InputType!]) { arrayArg(i: $var) }`, coerce.DefaultScalar)
			_, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, map[string]interface{}{
				"var": "hello",
			})
			require.EqualError(t, gerr, "input: variable.var must be an array")
		})

		t.Run("defaults", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: [InputType!] = [{name: "foo"}]) { arrayArg(i: $var) }`, coerce.DefaultScalar)
			vars, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, nil)
			require.Nil(t, gerr)
			require.EqualValues(t, []interface{}{map[string]interface{}{
				"name": "foo",
			}}, vars["var"])
		})

		t.Run("valid value", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: [InputType!]) { arrayArg(i: $var) }`, coerce.DefaultScalar)
			vars, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, map[string]interface{}{
				"var": []interface{}{map[string]interface{}{
					"name": "foo",
				}},
			})
			require.Nil(t, gerr)
			require.EqualValues(t, []interface{}{map[string]interface{}{
				"name": "foo",
			}}, vars["var"])
		})

		t.Run("null element value", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: [InputType!]) { arrayArg(i: $var) }`, coerce.DefaultScalar)
			_, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, map[string]interface{}{
				"var": []interface{}{nil},
			})
			require.EqualError(t, gerr, "input: variable.var[0] cannot be null")
		})

		t.Run("missing required values", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: [InputType!]) { arrayArg(i: $var) }`, coerce.DefaultScalar)
			_, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, map[string]interface{}{
				"var": []interface{}{map[string]interface{}{}},
			})
			require.EqualError(t, gerr, "input: variable.var[0].name must be defined")
		})
	})

	t.Run("Scalars", func(t *testing.T) {
		t.Run("String -> String", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: String!) { stringArg(i: $var) }`, coerce.DefaultScalar)
			vars, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, map[string]interface{}{
				"var": "asdf",
			})
			require.Nil(t, gerr)
			require.EqualValues(t, "asdf", vars["var"])
		})

		t.Run("Int -> String", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: String!) { stringArg(i: $var) }`, coerce.DefaultScalar)
			_, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, map[string]interface{}{
				"var": 1,
			})
			require.EqualError(t, gerr, "input: variable.var int cannot be coerced to String")
		})

		t.Run("Nil -> String", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: String!) { stringArg(i: $var) }`, coerce.DefaultScalar)
			_, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, map[string]interface{}{
				"var": nil,
			})
			require.EqualError(t, gerr, "input: variable.var cannot be null")
		})

		t.Run("Bool -> Int", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: Int!) { intArg(i: $var) }`, coerce.DefaultScalar)
			_, gerr := coerce.VariableValues(schema, q.Operations.ForName(""), coerce.DefaultScalar, map[string]interface{}{
				"var": true,
			})
			require.EqualError(t, gerr, "input: variable.var bool cannot be coerced to Int")
		})
	})
}

func mustReadFile(name string) string {
	src, err := ioutil.ReadFile(name)
	if err != nil {
		panic(err)
	}

	return string(src)
}
