package ast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArg2Map(t *testing.T) {
	defs := ArgumentDefinitionList{
		{
			Name:         "A",
			Type:         NamedType("String", nil),
			DefaultValue: &Value{Kind: StringValue, Raw: "defaultA"},
		},
		{Name: "B", Type: NamedType("String", nil)},
	}

	t.Run("defaults", func(t *testing.T) {
		args, err := arg2map(defs, ArgumentList{}, nil)
		require.NoError(t, err)
		require.Equal(t, "defaultA", args["A"])
		require.NotContains(t, args, "B")
	})

	t.Run("values", func(t *testing.T) {
		args, err := arg2map(defs, ArgumentList{
			{Name: "A", Value: &Value{Kind: StringValue, Raw: "valA"}},
			{Name: "B", Value: &Value{Kind: StringValue, Raw: "valB"}},
		}, nil)
		require.NoError(t, err)
		require.Equal(t, "valA", args["A"])
		require.Equal(t, "valB", args["B"])
	})

	t.Run("nulls", func(t *testing.T) {
		args, err := arg2map(defs, ArgumentList{
			{Name: "A", Value: &Value{Kind: NullValue}},
			{Name: "B", Value: &Value{Kind: NullValue}},
		}, nil)
		require.NoError(t, err)
		require.Nil(t, args["A"])
		require.Nil(t, args["B"])
		require.Contains(t, args, "A")
		require.Contains(t, args, "B")
	})

	t.Run("undefined variables", func(t *testing.T) {
		args, err := arg2map(defs, ArgumentList{
			{Name: "A", Value: &Value{Kind: Variable, Raw: "VarA"}},
			{Name: "B", Value: &Value{Kind: Variable, Raw: "VarB"}},
		}, map[string]any{})
		require.NoError(t, err)
		require.Equal(t, "defaultA", args["A"])
		require.NotContains(t, args, "B")
	})

	t.Run("nil variables", func(t *testing.T) {
		args, err := arg2map(defs, ArgumentList{
			{Name: "A", Value: &Value{Kind: Variable, Raw: "VarA"}},
			{Name: "B", Value: &Value{Kind: Variable, Raw: "VarB"}},
		}, map[string]any{
			"VarA": nil,
			"VarB": nil,
		})
		require.NoError(t, err)
		require.Nil(t, args["A"])
		require.Nil(t, args["B"])
		require.Contains(t, args, "A")
		require.Contains(t, args, "B")
	})

	t.Run("defined variables", func(t *testing.T) {
		args, err := arg2map(defs, ArgumentList{
			{Name: "A", Value: &Value{Kind: Variable, Raw: "VarA"}},
			{Name: "B", Value: &Value{Kind: Variable, Raw: "VarB"}},
		}, map[string]any{
			"VarA": "varvalA",
			"VarB": "varvalB",
		})
		require.NoError(t, err)
		require.Equal(t, "varvalA", args["A"])
		require.Equal(t, "varvalB", args["B"])
	})

	t.Run("invalid argument value", func(t *testing.T) {
		intDefs := ArgumentDefinitionList{
			{Name: "A", Type: NamedType("Int", nil)},
		}
		args, err := arg2map(intDefs, ArgumentList{
			{Name: "A", Value: &Value{Kind: IntValue, Raw: "99999999999999999999"}},
		}, nil)
		require.Error(t, err)
		require.Nil(t, args)
	})

	t.Run("invalid default value", func(t *testing.T) {
		intDefs := ArgumentDefinitionList{
			{
				Name:         "A",
				Type:         NamedType("Int", nil),
				DefaultValue: &Value{Kind: IntValue, Raw: "99999999999999999999"},
			},
		}
		args, err := arg2map(intDefs, ArgumentList{}, nil)
		require.Error(t, err)
		require.Nil(t, args)
	})

	t.Run("field ArgumentMap and TryArgumentMap", func(t *testing.T) {
		var noDef Field
		require.Nil(t, noDef.ArgumentMap(nil))
		m, err := noDef.TryArgumentMap(nil)
		require.NoError(t, err)
		require.Nil(t, m)

		valid := Field{
			Definition: &FieldDefinition{Arguments: defs},
			Arguments: ArgumentList{
				{Name: "A", Value: &Value{Kind: StringValue, Raw: "valA"}},
			},
		}
		require.Equal(t, map[string]any{"A": "valA"}, valid.ArgumentMap(nil))
		m, err = valid.TryArgumentMap(nil)
		require.NoError(t, err)
		require.Equal(t, map[string]any{"A": "valA"}, m)

		invalid := Field{
			Definition: &FieldDefinition{
				Arguments: ArgumentDefinitionList{{Name: "A", Type: NamedType("Int", nil)}},
			},
			Arguments: ArgumentList{
				{Name: "A", Value: &Value{Kind: IntValue, Raw: "99999999999999999999"}},
			},
		}
		m, err = invalid.TryArgumentMap(nil)
		require.Error(t, err)
		require.Nil(t, m)
		require.Panics(t, func() {
			invalid.ArgumentMap(nil)
		})
	})

	t.Run("directive ArgumentMap and TryArgumentMap", func(t *testing.T) {
		var noDef Directive
		require.Nil(t, noDef.ArgumentMap(nil))
		m, err := noDef.TryArgumentMap(nil)
		require.NoError(t, err)
		require.Nil(t, m)

		valid := Directive{
			Definition: &DirectiveDefinition{Arguments: defs},
			Arguments: ArgumentList{
				{Name: "A", Value: &Value{Kind: StringValue, Raw: "valA"}},
			},
		}
		require.Equal(t, map[string]any{"A": "valA"}, valid.ArgumentMap(nil))
		m, err = valid.TryArgumentMap(nil)
		require.NoError(t, err)
		require.Equal(t, map[string]any{"A": "valA"}, m)

		invalid := Directive{
			Definition: &DirectiveDefinition{
				Arguments: ArgumentDefinitionList{{Name: "A", Type: NamedType("Int", nil)}},
			},
			Arguments: ArgumentList{
				{Name: "A", Value: &Value{Kind: IntValue, Raw: "99999999999999999999"}},
			},
		}
		m, err = invalid.TryArgumentMap(nil)
		require.Error(t, err)
		require.Nil(t, m)
		require.Panics(t, func() {
			invalid.ArgumentMap(nil)
		})
	})
}
