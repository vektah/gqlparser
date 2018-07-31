package coerce

import (
	"reflect"

	"fmt"

	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/gqlerror"
)

// VariableValues coerces and validates variable values
func VariableValues(schema *ast.Schema, op *ast.OperationDefinition, coerceScalar ScalarFunc, variables map[string]interface{}) (map[string]interface{}, *gqlerror.Error) {
	coercedVars := map[string]interface{}{}

	validator := validator{
		path:         []interface{}{"variable"},
		schema:       schema,
		coerceScalar: coerceScalar,
	}

	for _, v := range op.VariableDefinitions {
		validator.path = append(validator.path, v.Variable)

		if !v.Definition.IsInputType() {
			return nil, gqlerror.ErrorPathf(validator.path, "must an input type")
		}

		val, hasValue := variables[v.Variable]
		if !hasValue {
			if v.DefaultValue != nil {
				var err error
				val, err = v.DefaultValue.Value(nil)
				if err != nil {
					return nil, gqlerror.WrapPath(validator.path, err)
				}
				hasValue = true
			} else if v.Type.NonNull {
				return nil, gqlerror.ErrorPathf(validator.path, "must be defined")
			}
		}

		rv := reflect.ValueOf(val)
		if v.Type.NonNull && val == nil {
			return nil, gqlerror.ErrorPathf(validator.path, "cannot be null")
		}

		if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
			rv = rv.Elem()
		}

		if err := validator.validateVarType(v.Type, rv); err != nil {
			return nil, err
		}

		if hasValue {
			coercedVars[v.Variable] = val
		}

		validator.path = validator.path[0 : len(validator.path)-1]
	}

	return coercedVars, nil
}

type validator struct {
	path         []interface{}
	schema       *ast.Schema
	coerceScalar ScalarFunc
}

func (v *validator) validateVarType(typ *ast.Type, val reflect.Value) *gqlerror.Error {
	if typ.Elem != nil {
		if val.Kind() != reflect.Slice {
			return gqlerror.ErrorPathf(v.path, "must be an array")
		}

		for i := 0; i < val.Len(); i++ {
			v.path = append(v.path, i)
			field := val.Index(i)

			if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
				if typ.Elem.NonNull && field.IsNil() {
					return gqlerror.ErrorPathf(v.path, "cannot be null")
				}
				field = field.Elem()
			}

			if err := v.validateVarType(typ.Elem, field); err != nil {
				return err
			}

			v.path = v.path[0 : len(v.path)-1]
		}

		return nil
	}

	def := v.schema.Types[typ.NamedType]
	if def == nil {
		panic(fmt.Errorf("missing def for %s", typ.NamedType))
	}

	switch def.Kind {
	case ast.Scalar, ast.Enum:
		_, err := v.coerceScalar(typ, def, val.Interface())
		if err != nil {
			return gqlerror.ErrorPathf(v.path, "%s cannot be coerced to %s", val.Kind().String(), def.Name)
		}
	case ast.InputObject:
		if val.Kind() != reflect.Map {
			return gqlerror.ErrorPathf(v.path, "must be a %s", def.Name)
		}

		// check for unknown fields
		for _, name := range val.MapKeys() {
			val.MapIndex(name)
			fieldDef := def.Fields.ForName(name.String())
			v.path = append(v.path, name)

			if fieldDef == nil {
				return gqlerror.ErrorPathf(v.path, "unknown field")
			}
			v.path = v.path[0 : len(v.path)-1]
		}

		for _, fieldDef := range def.Fields {
			v.path = append(v.path, fieldDef.Name)

			field := val.MapIndex(reflect.ValueOf(fieldDef.Name))
			if !field.IsValid() {
				if fieldDef.Type.NonNull {
					return gqlerror.ErrorPathf(v.path, "must be defined")
				}
				continue
			}

			if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
				if typ.NonNull && field.IsNil() {
					return gqlerror.ErrorPathf(v.path, "cannot be null")
				}
				field = field.Elem()
			}

			err := v.validateVarType(fieldDef.Type, field)
			if err != nil {
				return err
			}

			v.path = v.path[0 : len(v.path)-1]
		}
	default:
		panic(fmt.Errorf("unsupported type %s", def.Kind))
	}

	return nil
}
