package validator

import (
	"fmt"
	"strconv"

	"github.com/vektah/gqlparser/ast"
)

var UnexpectedType = fmt.Errorf("Unexpected Type")

type CoerceInputScalarFunc func(typeName string, kind ast.ValueKind, value string) (interface{}, error)

// CoerceScalar implements the coercion rules set out in
// http://facebook.github.io/graphql/June2018/#sec-Scalars
func DefaultInputCoercion(typeName string, kind ast.ValueKind, value string) (interface{}, error) {
	switch typeName {
	case "Int":
		return strconv.ParseInt(value, 10, 64)
	case "Float":
		return strconv.ParseFloat(value, 64)
	case "String":
		return value, nil
	case "Boolean":
		return value == "true", nil
	case "ID":
		return value, nil
	default:
		// custom scalars will pass through as strings
		return value, nil
	}
}

func CoerceValue(v *ast.Value, vars map[string]interface{}, inputFunc CoerceInputScalarFunc) (interface{}, error) {
	// variables have already been coerced at this point so it can just be returned
	if v.Kind == ast.Variable {
		return vars[v.Raw], nil
	}

	if v.Kind == ast.NullValue {
		return nil, nil
	}

	if v.ExpectedType.Elem != nil {
		if v.Kind != ast.ListValue {
			return nil, UnexpectedType
		}
		var val []interface{}
		for _, elem := range v.Children {
			elemVal, err := CoerceValue(elem.Value, vars, inputFunc)
			if err != nil {
				return val, err
			}
			val = append(val, elemVal)
		}
		return val, nil
	}

	if v.Definition.Kind == ast.InputObject {
		val := map[string]interface{}{}
		for _, elem := range v.Children {
			elemVal, err := CoerceValue(elem.Value, vars, inputFunc)
			if err != nil {
				return val, err
			}
			val[elem.Name] = elemVal
		}
		return val, nil
	}

	return inputFunc(v.Definition.Name, v.Kind, v.Raw)
}
