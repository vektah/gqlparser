package validator

import (
	"reflect"
	"strconv"
	"strings"

	"fmt"

	"github.com/dgraph-io/gqlparser/ast"
	"github.com/dgraph-io/gqlparser/gqlerror"
)

var UnexpectedType = fmt.Errorf("Unexpected Type")

func Variables(schema *ast.Schema, op *ast.OperationDefinition, variables map[string]interface{}) (map[string]*ast.Value, *gqlerror.Error) {
	coercedVars := map[string]*ast.Value{}

	validator := varValidator{
		path:   []interface{}{"variable"},
		schema: schema,
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

		if hasValue {
			if val == nil {
				if v.Type.NonNull {
					return nil, gqlerror.ErrorPathf(validator.path, "cannot be null")
				}
				coercedVars[v.Variable] = nil
			} else {
				rv := reflect.ValueOf(val)
				if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
					rv = rv.Elem()
				}

				vars, err := validator.getVars(v.Type, rv)
				if err != nil {
					return nil, err
				}

				coercedVars[v.Variable] = vars
			}
		}

		validator.path = validator.path[0 : len(validator.path)-1]
	}

	return coercedVars, nil
}

// VariableValues coerces and validates variable values
func VariableValues(schema *ast.Schema, op *ast.OperationDefinition, variables map[string]interface{}) (map[string]interface{}, *gqlerror.Error) {
	coercedVars := map[string]interface{}{}

	validator := varValidator{
		path:   []interface{}{"variable"},
		schema: schema,
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

		if hasValue {
			if val == nil {
				if v.Type.NonNull {
					return nil, gqlerror.ErrorPathf(validator.path, "cannot be null")
				}
				coercedVars[v.Variable] = nil
			} else {
				rv := reflect.ValueOf(val)
				if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
					rv = rv.Elem()
				}

				if err := validator.validateVarType(v.Type, rv); err != nil {
					return nil, err
				}

				coercedVars[v.Variable] = val
			}
		}

		validator.path = validator.path[0 : len(validator.path)-1]
	}

	return coercedVars, nil
}

type varValidator struct {
	path   []interface{}
	schema *ast.Schema
}

func (v *varValidator) getVars(typ *ast.Type, val reflect.Value) (*ast.Value, *gqlerror.Error) {
	currentPath := v.path
	resetPath := func() {
		v.path = currentPath
	}
	defer resetPath()

	value := &ast.Value{
		Raw:      val.String(),
		Children: make(ast.ChildValueList, 0, 0),
	}

	if typ.Elem != nil {
		if val.Kind() != reflect.Slice {
			return nil, gqlerror.ErrorPathf(v.path, "must be an array")
		}

		value.Kind = ast.ListValue

		for i := 0; i < val.Len(); i++ {
			resetPath()
			v.path = append(v.path, i)
			field := val.Index(i)

			if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
				if typ.Elem.NonNull && field.IsNil() {
					return nil, gqlerror.ErrorPathf(v.path, "cannot be null")
				}
				field = field.Elem()
			}

			vars, err := v.getVars(typ.Elem, field)
			if err != nil {
				return nil, err
			}

			value.Children = append(value.Children, &ast.ChildValue{
				Name:  strconv.Itoa(i),
				Value: vars,
			})

			// typ.Elem is of type [typ], so it doesn't exist in v.schema.Types
			//TODO find Definition from typ.Elem instead
			value.Definition = vars.Definition
		}

		return value, nil
	}

	if !typ.NonNull && !val.IsValid() {
		// If the type is not null and we got a invalid value namely null/nil, then it's valid
		return nil, nil
	}
	def := v.schema.Types[typ.NamedType]
	if def == nil {
		panic(fmt.Errorf("missing def for %s", typ.NamedType))
	}
	value.Definition = def
	value.Raw = val.String()

	switch def.Kind {
	case ast.Enum:
		kind := val.Type().Kind()
		value.Kind = ast.EnumValue
		if kind != reflect.Int && kind != reflect.Int32 && kind != reflect.Int64 && kind != reflect.String {
			return nil, gqlerror.ErrorPathf(v.path, "enums must be ints or strings")
		}
		isValidEnum := false
		for _, enumVal := range def.EnumValues {
			if strings.EqualFold(val.String(), enumVal.Name) {
				isValidEnum = true
			}
		}
		if !isValidEnum {
			return nil, gqlerror.ErrorPathf(v.path, "%s is not a valid %s", val.String(), def.Name)
		}
		return value, nil
	case ast.Scalar:
		kind := val.Type().Kind()
		switch typ.NamedType {
		case "Int":
			if kind == reflect.String || kind == reflect.Int || kind == reflect.Int32 || kind == reflect.Int64 {
				value.Kind = ast.IntValue
				return value, nil
			}
		case "Float":
			if kind == reflect.String || kind == reflect.Float32 || kind == reflect.Float64 || kind == reflect.Int || kind == reflect.Int32 || kind == reflect.Int64 {
				value.Kind = ast.FloatValue
				return value, nil
			}
		case "String":
			if kind == reflect.String {
				value.Kind = ast.StringValue
				return value, nil
			}

		case "Boolean":
			if kind == reflect.Bool {
				value.Kind = ast.BooleanValue
				return value, nil
			}

		case "ID":
			if kind == reflect.Int || kind == reflect.Int32 || kind == reflect.Int64 || kind == reflect.String {
				value.Kind = ast.IntValue
			}
		default:
			// assume custom scalars are ok
			value.Kind = ast.StringValue
			return value, nil
		}
		return nil, gqlerror.ErrorPathf(v.path, "cannot use %s as %s", kind.String(), typ.NamedType)
	case ast.InputObject:
		if val.Kind() != reflect.Map {
			return nil, gqlerror.ErrorPathf(v.path, "must be a %s", def.Name)
		}

		value.Kind = ast.ObjectValue

		// check for unknown fields
		for _, name := range val.MapKeys() {
			val.MapIndex(name)
			fieldDef := def.Fields.ForName(name.String())
			resetPath()
			v.path = append(v.path, name.String())

			if fieldDef == nil {
				return nil, gqlerror.ErrorPathf(v.path, "unknown field")
			}
		}

		for _, fieldDef := range def.Fields {
			resetPath()
			v.path = append(v.path, fieldDef.Name)

			field := val.MapIndex(reflect.ValueOf(fieldDef.Name))
			if !field.IsValid() {
				if fieldDef.Type.NonNull {
					return nil, gqlerror.ErrorPathf(v.path, "must be defined")
				}
				continue
			}

			if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
				if fieldDef.Type.NonNull && field.IsNil() {
					return nil, gqlerror.ErrorPathf(v.path, "cannot be null")
				}
				//allow null object field and skip it
				if !fieldDef.Type.NonNull && field.IsNil() {
					continue
				}
				field = field.Elem()
			}

			val, err := v.getVars(fieldDef.Type, field)
			if err != nil {
				return nil, err
			}
			value.Children = append(value.Children, &ast.ChildValue{
				Name:  fieldDef.Name,
				Value: val,
			})
		}
	default:
		panic(fmt.Errorf("unsupported type %s", def.Kind))
	}

	return value, nil
}

func (v *varValidator) validateVarType(typ *ast.Type, val reflect.Value) *gqlerror.Error {
	currentPath := v.path
	resetPath := func() {
		v.path = currentPath
	}
	defer resetPath()

	if typ.Elem != nil {
		if val.Kind() != reflect.Slice {
			return gqlerror.ErrorPathf(v.path, "must be an array")
		}

		for i := 0; i < val.Len(); i++ {
			resetPath()
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
		}

		return nil
	}

	def := v.schema.Types[typ.NamedType]
	if def == nil {
		panic(fmt.Errorf("missing def for %s", typ.NamedType))
	}

	if !typ.NonNull && !val.IsValid() {
		// If the type is not null and we got a invalid value namely null/nil, then it's valid
		return nil
	}

	switch def.Kind {
	case ast.Enum:
		kind := val.Type().Kind()
		if kind != reflect.Int && kind != reflect.Int32 && kind != reflect.Int64 && kind != reflect.String {
			return gqlerror.ErrorPathf(v.path, "enums must be ints or strings")
		}
		isValidEnum := false
		for _, enumVal := range def.EnumValues {
			if strings.EqualFold(val.String(), enumVal.Name) {
				isValidEnum = true
			}
		}
		if !isValidEnum {
			return gqlerror.ErrorPathf(v.path, "%s is not a valid %s", val.String(), def.Name)
		}
		return nil
	case ast.Scalar:
		kind := val.Type().Kind()
		switch typ.NamedType {
		case "Int":
			if kind == reflect.String || kind == reflect.Int || kind == reflect.Int32 || kind == reflect.Int64 {
				return nil
			}
		case "Float":
			if kind == reflect.String || kind == reflect.Float32 || kind == reflect.Float64 || kind == reflect.Int || kind == reflect.Int32 || kind == reflect.Int64 {
				return nil
			}
		case "String":
			if kind == reflect.String {
				return nil
			}

		case "Boolean":
			if kind == reflect.Bool {
				return nil
			}

		case "ID":
			if kind == reflect.Int || kind == reflect.Int32 || kind == reflect.Int64 || kind == reflect.String {
				return nil
			}
		default:
			// assume custom scalars are ok
			return nil
		}
		return gqlerror.ErrorPathf(v.path, "cannot use %s as %s", kind.String(), typ.NamedType)
	case ast.InputObject:
		if val.Kind() != reflect.Map {
			return gqlerror.ErrorPathf(v.path, "must be a %s", def.Name)
		}

		// check for unknown fields
		for _, name := range val.MapKeys() {
			val.MapIndex(name)
			fieldDef := def.Fields.ForName(name.String())
			resetPath()
			v.path = append(v.path, name.String())

			if fieldDef == nil {
				return gqlerror.ErrorPathf(v.path, "unknown field")
			}
		}

		for _, fieldDef := range def.Fields {
			resetPath()
			v.path = append(v.path, fieldDef.Name)

			field := val.MapIndex(reflect.ValueOf(fieldDef.Name))
			if !field.IsValid() {
				if fieldDef.Type.NonNull {
					return gqlerror.ErrorPathf(v.path, "must be defined")
				}
				continue
			}

			if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
				if fieldDef.Type.NonNull && field.IsNil() {
					return gqlerror.ErrorPathf(v.path, "cannot be null")
				}
				//allow null object field and skip it
				if !fieldDef.Type.NonNull && field.IsNil() {
					continue
				}
				field = field.Elem()
			}

			err := v.validateVarType(fieldDef.Type, field)
			if err != nil {
				return err
			}
		}
	default:
		panic(fmt.Errorf("unsupported type %s", def.Kind))
	}

	return nil
}
