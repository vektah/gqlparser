package coerce

import (
	"reflect"

	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/gqlerror"
)

// FieldArguments coerces all arguments for a given field into a map, taking into account defaults
func FieldArguments(field *ast.Field, variables map[string]interface{}, coerceScalar ScalarFunc) (map[string]interface{}, *gqlerror.Error) {
	coercedValues := map[string]interface{}{}

	for _, arg := range field.Definition.Arguments {
		argumentValue := field.Arguments.ForName(arg.Name)
		var hasValue bool
		var value interface{}
		var err error
		if argumentValue != nil {
			if argumentValue.Value.Kind == ast.Variable {
				value, hasValue = variables[argumentValue.Value.Raw]
			} else {
				value, err = argumentValue.Value.Value(variables)
				if err != nil {
					return nil, gqlerror.ErrorPosf(argumentValue.Position, err.Error())
				}
				hasValue = true
			}
		}

		if !hasValue && arg.DefaultValue != nil {
			value, err = arg.DefaultValue.Value(nil)
			hasValue = true
			if err != nil {
				return nil, gqlerror.ErrorPosf(field.Position, err.Error())
			}
		} else if arg.Type.NonNull && (!hasValue || value == nil) {
			return nil, gqlerror.ErrorPosf(field.Position, "argument %s must be provided", arg.Name)
		}

		if hasValue {
			rv := reflect.ValueOf(&value).Elem()
			cv, err := coerceValue(argumentValue.Value.ExpectedType, argumentValue.Value.Definition, coerceScalar, rv)
			if err != nil {
				return nil, gqlerror.ErrorPosf(argumentValue.Value.Position, err.Error())
			}
			coercedValues[arg.Name] = cv.Interface()
		}
	}
	return coercedValues, nil
}

func coerceValue(expected *ast.Type, def *ast.Definition, coerceScalar ScalarFunc, v reflect.Value) (reflect.Value, error) {
	if v.IsNil() {
		return v, nil
	}
	if expected == nil {
		return reflect.ValueOf(nil), nil
	}
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	if expected.Elem != nil {
		// Insane single element -> list conversion. see http://facebook.github.io/graphql/June2018/#sec-Type-System.List
		if v.Kind() != reflect.Slice {
			elemVal, err := coerceValue(expected.Elem, def, coerceScalar, v)
			if err != nil {
				return reflect.ValueOf(nil), err
			}
			return reflect.ValueOf(&[]interface{}{elemVal.Interface()}).Elem(), nil
		}

		for i := 0; i < v.Len(); i++ {
			field := v.Index(i)
			val, err := coerceValue(expected.Elem, def, coerceScalar, field)
			if err != nil {
				return val, err
			}
			field.Set(val)
		}
		return v, nil
	}

	if def.Kind == ast.InputObject {
		if v.Kind() != reflect.Map {
			return reflect.ValueOf(nil), UnexpectedType
		}

		for _, key := range v.MapKeys() {
			field := v.MapIndex(key)
			fieldDef := def.Fields.ForName(key.String())

			val, err := coerceValue(fieldDef.Type, fieldDef.ResultDefinition, coerceScalar, field)
			if err != nil {
				return val, err
			}
			v.SetMapIndex(key, val)
		}
		return v, nil
	}

	coerced, err := coerceScalar(expected, def, v.Interface())
	return reflect.ValueOf(&coerced).Elem(), err
}
