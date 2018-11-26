package validator_test

import (
	"io/ioutil"
	"testing"

	"encoding/json"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/validator"
)

func TestValidateVars(t *testing.T) {
	schema := gqlparser.MustLoadSchema(&ast.Source{
		Name:  "vars.graphql",
		Input: mustReadFile("./testdata/vars.graphql"),
	})

	t.Run("undefined variable", func(t *testing.T) {
		t.Run("without default", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query($id: Int!) { intArg(i: $id) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), nil)
			require.EqualError(t, gerr, "input: variable.id must be defined")
		})

		t.Run("nil in required value", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query($id: Int!) { intArg(i: $id) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"id": nil,
			})
			require.EqualError(t, gerr, "input: variable.id cannot be null")
		})

		t.Run("with default", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query($id: Int! = 1) { intArg(i: $id) }`)
			vars, gerr := validator.VariableValues(schema, q.Operations.ForName(""), nil)
			require.Nil(t, gerr)
			require.EqualValues(t, 1, vars["id"])
		})

		t.Run("with union", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query($id: Int! = 1) { intArg(i: $id) }`)
			vars, gerr := validator.VariableValues(schema, q.Operations.ForName(""), nil)
			require.Nil(t, gerr)
			require.EqualValues(t, 1, vars["id"])
		})
	})

	t.Run("input object", func(t *testing.T) {
		t.Run("non object", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var": "hello",
			})
			require.EqualError(t, gerr, "input: variable.var must be a InputType")
		})

		t.Run("defaults", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType! = {name: "foo"}) { structArg(i: $var) }`)
			vars, gerr := validator.VariableValues(schema, q.Operations.ForName(""), nil)
			require.Nil(t, gerr)
			require.EqualValues(t, map[string]interface{}{"name": "foo"}, vars["var"])
		})

		t.Run("valid value", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`)
			vars, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var": map[string]interface{}{
					"name": "foobar",
				},
			})
			require.Nil(t, gerr)
			require.EqualValues(t, map[string]interface{}{"name": "foobar"}, vars["var"])
		})

		t.Run("null object field", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`)
			vars, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var": map[string]interface{}{
					"name":     "foobar",
					"nullName": nil,
				},
			})
			require.Nil(t, gerr)
			require.EqualValues(t, map[string]interface{}{"name": "foobar", "nullName": nil}, vars["var"])
		})

		t.Run("missing required values", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var": map[string]interface{}{},
			})
			require.EqualError(t, gerr, "input: variable.var.name must be defined")
		})

		t.Run("null required field", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var": map[string]interface{}{
					"name": nil,
				},
			})
			require.EqualError(t, gerr, "input: variable.var.name cannot be null")
		})

		t.Run("null embedded input object", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var": map[string]interface{}{
					"name":         "foo",
					"nullEmbedded": nil,
				},
			})
			require.Nil(t, gerr)
		})

		t.Run("unknown field", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
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
			q := gqlparser.MustLoadQuery(schema, `query foo($var: [InputType!]) { arrayArg(i: $var) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var": "hello",
			})
			require.EqualError(t, gerr, "input: variable.var must be an array")
		})

		t.Run("defaults", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: [InputType!] = [{name: "foo"}]) { arrayArg(i: $var) }`)
			vars, gerr := validator.VariableValues(schema, q.Operations.ForName(""), nil)
			require.Nil(t, gerr)
			require.EqualValues(t, []interface{}{map[string]interface{}{
				"name": "foo",
			}}, vars["var"])
		})

		t.Run("valid value", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: [InputType!]) { arrayArg(i: $var) }`)
			vars, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
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
			q := gqlparser.MustLoadQuery(schema, `query foo($var: [InputType!]) { arrayArg(i: $var) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var": []interface{}{nil},
			})
			require.EqualError(t, gerr, "input: variable.var[0] cannot be null")
		})

		t.Run("missing required values", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: [InputType!]) { arrayArg(i: $var) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var": []interface{}{map[string]interface{}{}},
			})
			require.EqualError(t, gerr, "input: variable.var[0].name must be defined")
		})
		t.Run("invalid variable paths", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var1: InputType!, $var2: InputType!) { a:structArg(i: $var1) b:structArg(i: $var2) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var1": map[string]interface{}{
					"name": "foobar",
				},
				"var2": map[string]interface{}{
					"nullName": "foobar",
				},
			})
			require.EqualError(t, gerr, "input: variable.var2.name must be defined")
		})
	})

	t.Run("Scalars", func(t *testing.T) {
		t.Run("String -> String", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: String!) { stringArg(i: $var) }`)
			vars, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var": "asdf",
			})
			require.Nil(t, gerr)
			require.EqualValues(t, "asdf", vars["var"])
		})

		t.Run("Int -> String", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: String!) { stringArg(i: $var) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var": 1,
			})
			require.EqualError(t, gerr, "input: variable.var cannot use int as String")
		})

		t.Run("Nil -> String", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: String!) { stringArg(i: $var) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var": nil,
			})
			require.EqualError(t, gerr, "input: variable.var cannot be null")
		})

		t.Run("Undefined -> String!", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: String!) { stringArg(i: $var) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), nil)
			require.EqualError(t, gerr, "input: variable.var must be defined")
		})

		t.Run("Undefined -> Int", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: Int) { optionalIntArg(i: $var) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), nil)
			require.Nil(t, gerr)
		})

		t.Run("Json Number -> Int", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: Int) { optionalIntArg(i: $var) }`)
			vars, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var": json.Number("10"),
			})
			require.Nil(t, gerr)
			require.Equal(t, json.Number("10"), vars["var"])
		})

		t.Run("Nil -> Int", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: Int) { optionalIntArg(i: $var) }`)
			vars, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var": nil,
			})
			require.Nil(t, gerr)
			require.Equal(t, nil, vars["var"])
		})

		t.Run("Bool -> Int", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: Int!) { intArg(i: $var) }`)
			_, gerr := validator.VariableValues(schema, q.Operations.ForName(""), map[string]interface{}{
				"var": true,
			})
			require.EqualError(t, gerr, "input: variable.var cannot use bool as Int")
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
