package variable

import (
	"fmt"

	"github.com/vektah/gqlparser/ast"
)

var UnexpectedType = fmt.Errorf("Unexpected Type")

func (b *Bag) CoerceValue(v *ast.Value) (interface{}, error) {
	// variables have already been coerced at this point so it can just be returned
	if v.Kind == ast.Variable {
		return b.coercedVars[v.Raw], nil
	}

	if v.Kind == ast.NullValue {
		return nil, nil
	}

	if v.ExpectedType == nil {
		return nil, nil
	}

	if v.ExpectedType.Elem != nil {
		// Implement insane single value -> list conversion. see http://facebook.github.io/graphql/June2018/#sec-Type-System.List
		if v.Kind != ast.ListValue {
			cpy := *v
			cpy.ExpectedType = v.ExpectedType.Elem
			elemVal, err := b.CoerceValue(&cpy)
			return []interface{}{elemVal}, err
		}

		var val []interface{}
		for _, elem := range v.Children {
			elemVal, err := b.CoerceValue(elem.Value)
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
			elemVal, err := b.CoerceValue(elem.Value)
			if err != nil {
				return val, err
			}
			val[elem.Name] = elemVal
		}
		return val, nil
	}

	goVal, err := b.Value(v)
	if err != nil {
		return nil, err
	}
	return b.CoerceScalar(v.ExpectedType, v.Definition, goVal)
}
