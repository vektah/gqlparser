package variable_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/variable"
)

func TestCoerceVars(t *testing.T) {
	schema := gqlparser.MustLoadSchema(&ast.Source{
		Name:  "vars.graphql",
		Input: mustReadFile("./testdata/vars.graphql"),
	})

	t.Run("undefined variable", func(t *testing.T) {
		t.Run("without default", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query($id: Int!) { intArg(i: $id) }`, variable.DefaultInputCoercion)
			_, gerr := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, nil)
			require.EqualError(t, gerr, "input: variable.id must be defined")
		})

		t.Run("nil in required value", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query($id: Int!) { intArg(i: $id) }`, variable.DefaultInputCoercion)
			_, gerr := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, map[string]interface{}{
				"id": nil,
			})
			require.EqualError(t, gerr, "input: variable.id cannot be null")
		})

		t.Run("with default", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query($id: Int! = 1) { intArg(i: $id) }`, variable.DefaultInputCoercion)
			vars, gerr := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, nil)
			require.Nil(t, gerr)
			require.EqualValues(t, 1, vars.Get("id"))
		})

		t.Run("with union", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query($id: Int! = 1) { intArg(i: $id) }`, variable.DefaultInputCoercion)
			vars, gerr := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, nil)
			require.Nil(t, gerr)
			require.EqualValues(t, 1, vars.Get("id"))
		})
	})

	t.Run("input object", func(t *testing.T) {
		t.Run("non object", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`, variable.DefaultInputCoercion)
			_, gerr := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, map[string]interface{}{
				"var": "hello",
			})
			require.EqualError(t, gerr, "input: variable.var must be a InputType")
		})

		t.Run("defaults", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType! = {name: "foo"}) { structArg(i: $var) }`, variable.DefaultInputCoercion)
			vars, gerr := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, nil)
			require.Nil(t, gerr)
			require.EqualValues(t, map[string]interface{}{"name": "foo"}, vars.Get("var"))
		})

		t.Run("valid value", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`, variable.DefaultInputCoercion)
			vars, gerr := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, map[string]interface{}{
				"var": map[string]interface{}{
					"name": "foobar",
				},
			})
			require.Nil(t, gerr)
			require.EqualValues(t, map[string]interface{}{"name": "foobar"}, vars.Get("var"))
		})

		t.Run("missing required values", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`, variable.DefaultInputCoercion)
			_, gerr := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, map[string]interface{}{
				"var": map[string]interface{}{},
			})
			require.EqualError(t, gerr, "input: variable.var.name must be defined")
		})

		t.Run("null required field", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`, variable.DefaultInputCoercion)
			_, gerr := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, map[string]interface{}{
				"var": map[string]interface{}{
					"name": nil,
				},
			})
			require.EqualError(t, gerr, "input: variable.var.name cannot be null")
		})

		t.Run("unknown field", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: InputType!) { structArg(i: $var) }`, variable.DefaultInputCoercion)
			_, gerr := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, map[string]interface{}{
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
			q := gqlparser.MustLoadQuery(schema, `query foo($var: [InputType!]) { arrayArg(i: $var) }`, variable.DefaultInputCoercion)
			_, gerr := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, map[string]interface{}{
				"var": "hello",
			})
			require.EqualError(t, gerr, "input: variable.var must be an array")
		})

		t.Run("defaults", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: [InputType!] = [{name: "foo"}]) { arrayArg(i: $var) }`, variable.DefaultInputCoercion)
			vars, gerr := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, nil)
			require.Nil(t, gerr)
			require.EqualValues(t, []interface{}{map[string]interface{}{
				"name": "foo",
			}}, vars.Get("var"))
		})

		t.Run("valid value", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: [InputType!]) { arrayArg(i: $var) }`, variable.DefaultInputCoercion)
			vars, gerr := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, map[string]interface{}{
				"var": []interface{}{map[string]interface{}{
					"name": "foo",
				}},
			})
			require.Nil(t, gerr)
			require.EqualValues(t, []interface{}{map[string]interface{}{
				"name": "foo",
			}}, vars.Get("var"))
		})

		t.Run("null element value", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: [InputType!]) { arrayArg(i: $var) }`, variable.DefaultInputCoercion)
			_, gerr := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, map[string]interface{}{
				"var": []interface{}{nil},
			})
			require.EqualError(t, gerr, "input: variable.var[0] cannot be null")
		})

		t.Run("missing required values", func(t *testing.T) {
			q := gqlparser.MustLoadQuery(schema, `query foo($var: [InputType!]) { arrayArg(i: $var) }`, variable.DefaultInputCoercion)
			_, gerr := variable.NewBag(schema, q.Operations.ForName(""), variable.DefaultInputCoercion, map[string]interface{}{
				"var": []interface{}{map[string]interface{}{}},
			})
			require.EqualError(t, gerr, "input: variable.var[0].name must be defined")
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
