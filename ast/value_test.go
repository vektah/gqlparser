package ast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValueObject(t *testing.T) {
	t.Run("undefined variable", func(t *testing.T) {
		obj := &Value{Kind: ObjectValue, Children: ChildValueList{
			{Name: "field", Value: &Value{Kind: Variable, Raw: "Var"}},
		}}
		val, err := obj.Value(map[string]any{})
		require.NoError(t, err)
		// Treated as absent so the field's own default can be applied later.
		require.NotContains(t, val.(map[string]any), "field")
	})

	t.Run("undefined variable with default", func(t *testing.T) {
		// The variable is not supplied in vars, but its definition carries a
		// default value, so the field is present with that default.
		obj := &Value{Kind: ObjectValue, Children: ChildValueList{
			{Name: "field", Value: &Value{
				Kind: Variable,
				Raw:  "Var",
				VariableDefinition: &VariableDefinition{
					Variable:     "Var",
					DefaultValue: &Value{Kind: IntValue, Raw: "42"},
				},
			}},
		}}
		val, err := obj.Value(map[string]any{})
		require.NoError(t, err)
		m := val.(map[string]any)
		require.Contains(t, m, "field")
		require.Equal(t, int64(42), m["field"])
	})

	t.Run("explicit null", func(t *testing.T) {
		obj := &Value{Kind: ObjectValue, Children: ChildValueList{
			{Name: "field", Value: &Value{Kind: NullValue}},
		}}
		val, err := obj.Value(map[string]any{})
		require.NoError(t, err)
		m := val.(map[string]any)
		require.Contains(t, m, "field")
		require.Nil(t, m["field"])
	})
}
