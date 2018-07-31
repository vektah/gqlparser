package coerce

import (
	"strconv"

	"fmt"

	"github.com/vektah/gqlparser/ast"
)

var UnexpectedType = fmt.Errorf("Unexpected Type")

type ScalarFunc func(expected *ast.Type, def *ast.Definition, value interface{}) (interface{}, error)

// CoerceScalar implements the coercion rules set out in
// http://facebook.github.io/graphql/June2018/#sec-Scalars
func DefaultScalar(expected *ast.Type, def *ast.Definition, value interface{}) (interface{}, error) {
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
		if v, ok := value.(int); ok {
			return int64(v), nil
		}
		if v, ok := value.(int32); ok {
			return int64(v), nil
		}
		return nil, UnexpectedType
	case "Float":
		if v, ok := value.(int64); ok {
			return float64(v), nil
		}
		if v, ok := value.(int); ok {
			return float64(v), nil
		}
		if v, ok := value.(int32); ok {
			return float64(v), nil
		}
		if v, ok := value.(float64); ok {
			return v, nil
		}
		if v, ok := value.(float32); ok {
			return float64(v), nil
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
		if v, ok := value.(int); ok {
			return strconv.FormatInt(int64(v), 10), nil
		}
		if v, ok := value.(int32); ok {
			return strconv.FormatInt(int64(v), 10), nil
		}
		return nil, UnexpectedType
	default:
		// custom scalars will pass through
		return value, nil
	}
}
