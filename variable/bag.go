package variable

import (
	"strconv"

	"fmt"

	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/gqlerror"
)

type Bag struct {
	coercedVars  map[string]interface{}
	CoerceScalar CoerceInputScalarFunc
}

type CoerceInputScalarFunc func(expected *ast.Type, def *ast.Definition, value interface{}) (interface{}, error)

func NewBag(schema *ast.Schema, op *ast.OperationDefinition, inputFunc CoerceInputScalarFunc, variables map[string]interface{}) (*Bag, *gqlerror.Error) {
	b := &Bag{
		CoerceScalar: inputFunc,
	}

	return b, b.coerceVariableValues(schema, op, variables)
}

func NewEmptyBag(inputFunc CoerceInputScalarFunc) *Bag {
	return &Bag{
		CoerceScalar: inputFunc,
	}
}

func (b *Bag) Get(name string) interface{} {
	return b.coercedVars[name]
}

func (b *Bag) Has(name string) bool {
	_, has := b.coercedVars[name]
	return has
}

func (b *Bag) Value(v *ast.Value) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	switch v.Kind {
	case ast.Variable:
		if value, ok := b.coercedVars[v.Raw]; ok {
			return value, nil
		}
		if v.VariableDefinition != nil && v.VariableDefinition.DefaultValue != nil {
			return b.Value(v.VariableDefinition.DefaultValue)
		}
		return nil, nil
	case ast.IntValue:
		return strconv.ParseInt(v.Raw, 10, 64)
	case ast.FloatValue:
		return strconv.ParseFloat(v.Raw, 64)
	case ast.StringValue, ast.BlockValue, ast.EnumValue:
		return v.Raw, nil
	case ast.BooleanValue:
		return strconv.ParseBool(v.Raw)
	case ast.NullValue:
		return nil, nil
	case ast.ListValue:
		var val []interface{}
		for _, elem := range v.Children {
			elemVal, err := b.Value(elem.Value)
			if err != nil {
				return val, err
			}
			val = append(val, elemVal)
		}
		return val, nil
	case ast.ObjectValue:
		val := map[string]interface{}{}
		for _, elem := range v.Children {
			elemVal, err := b.Value(elem.Value)
			if err != nil {
				return val, err
			}
			val[elem.Name] = elemVal
		}
		return val, nil
	default:
		panic(fmt.Errorf("unknown value kind %d", v.Kind))
	}
}

// CoerceScalar implements the coercion rules set out in
// http://facebook.github.io/graphql/June2018/#sec-Scalars
func DefaultInputCoercion(expected *ast.Type, def *ast.Definition, value interface{}) (interface{}, error) {
	if def != nil && def.Kind == ast.Enum {
		if v, ok := value.(string); ok {
			return v, nil
		}
		return nil, UnexpectedType
	}
	switch expected.NamedType {
	case "Int":
		if v, ok := value.(int64); ok {
			return v, nil
		}
		return nil, UnexpectedType
	case "Float":
		if v, ok := value.(int64); ok {
			return float64(v), nil
		}
		if v, ok := value.(float64); ok {
			return v, nil
		}
		return nil, UnexpectedType
	case "String":
		if v, ok := value.(string); ok {
			return v, nil
		}

		return nil, UnexpectedType
	case "Boolean":
		if v, ok := value.(bool); ok {
			return v, nil
		}
		return nil, UnexpectedType
	case "ID":
		if v, ok := value.(string); ok {
			return v, nil
		}
		if v, ok := value.(int64); ok {
			return strconv.FormatInt(v, 10), nil
		}
		return nil, UnexpectedType
	default:
		// custom scalars will pass through
		return value, nil
	}
}
