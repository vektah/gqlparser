package ast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultValue(t *testing.T) {
	v := Value{
		Raw:  "foo",
		Kind: Variable,
		VariableDefinition: &VariableDefinition{
			DefaultValue: &Value{
				Raw:  "99",
				Kind: IntValue,
			},
		},
	}

	t.Run("returns variable value when provided", func(t *testing.T) {
		vars := make(map[string]interface{})
		vars["foo"] = int64(123)
		value, _ := v.Value(vars)
		require.Equal(t, int64(123), value)
	})

	t.Run("resolves default value when variable not provided", func(t *testing.T) {
		value, _ := v.Value(make(map[string]interface{}))
		require.Equal(t, int64(99), value)
	})

	t.Run("returns error when variable has no default", func(t *testing.T) {
		v := Value{Raw: "foo", Kind: Variable, VariableDefinition: &VariableDefinition{}}
		_, err := v.Value(make(map[string]interface{}))
		require.Error(t, err)
	})
}
