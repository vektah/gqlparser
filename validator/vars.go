package validator

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/gqlparser/v2/ast"
	"github.com/dgraph-io/gqlparser/v2/gqlerror"
	"reflect"
	"strconv"
	"strings"
)

// VariableValues coerces and validates variable values
func VariableValues(schema *ast.Schema, op *ast.OperationDefinition, variables map[string]interface{}) (map[string]interface{}, *gqlerror.Error) {
	coercedVars := map[string]interface{}{}

	validator := varValidator{
		path:   ast.Path{ast.PathName("variable")},
		schema: schema,
	}

	for _, v := range op.VariableDefinitions {
		validator.path = append(validator.path, ast.PathName(v.Variable))

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

				rval, err := validator.validateVarType(v.Type, rv)
				if err != nil {
					return nil, err
				}
				coercedVars[v.Variable] = rval.Interface()

			}
		}

		validator.path = validator.path[0 : len(validator.path)-1]
	}
	// cascade directive arguments validation
	if op.VariableDefinitions != nil {
		err := validator.cascadeDirectiveValidation(op, op.SelectionSet, variables)
		if err != nil {
			return nil, err
		}
	}
	return coercedVars, nil
}

func (v *varValidator) cascadeDirectiveValidation(op *ast.OperationDefinition, sel ast.SelectionSet, variables map[string]interface{}) *gqlerror.Error {

	for _, s := range sel {
		if f, ok := s.(*ast.Field); ok {
			cascadedir := f.Directives.ForName("cascade")
			if cascadedir == nil {
				continue
			}
			if len(cascadedir.Arguments) == 1 {
				if cascadedir.ParentDefinition == nil {
					return gqlerror.Errorf("Schema is not set yet. Please try after sometime.")
				}
			}
			if cascadedir.Arguments.ForName("fields") == nil {
				continue
			}
			fieldArg := cascadedir.Arguments.ForName("fields")
			isVariable := strings.HasPrefix(fieldArg.Value.String(), "$")
			if !isVariable {
				continue
			}

			varName := op.VariableDefinitions.ForName(cascadedir.Arguments.ForName("fields").Value.Raw)
			v.path = append(v.path, ast.PathName(varName.Variable))
			if cascadedir.ArgumentMap(variables)["fields"] == nil {
				return gqlerror.ErrorPathf(v.path, "variable %s not defined", varName.Variable)
			}

			variableVal := cascadedir.ArgumentMap(variables)["fields"].([]interface{})
			typFields := cascadedir.ParentDefinition.Fields
			typName := cascadedir.ParentDefinition.Name
			for _, val := range variableVal {
				if typFields.ForName(val.(string)) == nil {
					v.path = append(v.path, ast.PathName(val.(string)))
					return gqlerror.ErrorPathf(v.path, "Field `%s` is not present in type `%s`."+
						" You can only use fields which are in type `%s`", val, typName, typName)
				}
			}
			v.path = v.path[0 : len(v.path)-1]
			err := v.cascadeDirectiveValidation(op, f.SelectionSet, variables)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type varValidator struct {
	path   ast.Path
	schema *ast.Schema
}

func (v *varValidator) validateVarType(typ *ast.Type, val reflect.Value) (reflect.Value, *gqlerror.Error) {
	currentPath := v.path
	resetPath := func() {
		v.path = currentPath
	}
	defer resetPath()
	slc := make([]interface{}, 0)
	if typ.Elem != nil {
		if val.Kind() != reflect.Slice {
			// GraphQL spec says that non-null values should be coerced to an array when possible.
			// Hence if the value is not a slice, we create a slice and add val to it.
			if typ.Name() == "ID" && val.Type().Name() != "string" {
				val = val.Convert((reflect.ValueOf("string")).Type())
				slc = append(slc, val.String())
			} else {
				slc = append(slc, val.Interface())
			}
			val = reflect.ValueOf(slc)
		}
		slc = []interface{}{}
		for i := 0; i < val.Len(); i++ {
			resetPath()
			v.path = append(v.path, ast.PathIndex(i))
			field := val.Index(i)
			if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
				if typ.Elem.NonNull && field.IsNil() {
					return val, gqlerror.ErrorPathf(v.path, "cannot be null")
				}
				field = field.Elem()
			}
			cval, err := v.validateVarType(typ.Elem, field)
			if typ.Name() == "ID" {
				if val.Type().Name() != "string" {
					cval = cval.Convert((reflect.ValueOf("string")).Type())
				}
				slc = append(slc, cval.String())
			}
			if err != nil {
				return val, err
			}
		}
		if typ.Name() == "ID" {
			val = reflect.ValueOf(slc)
		}
		return val, nil
	}
	def := v.schema.Types[typ.NamedType]
	if def == nil {
		panic(fmt.Errorf("missing def for %s", typ.NamedType))
	}

	if !typ.NonNull && !val.IsValid() {
		// If the type is not null and we got a invalid value namely null/nil, then it's valid
		return val, nil
	}

	switch def.Kind {
	case ast.Enum:
		kind := val.Type().Kind()
		if kind != reflect.Int && kind != reflect.Int32 && kind != reflect.Int64 && kind != reflect.String {
			return val, gqlerror.ErrorPathf(v.path, "enums must be ints or strings")
		}
		isValidEnum := false
		for _, enumVal := range def.EnumValues {
			if strings.EqualFold(val.String(), enumVal.Name) {
				isValidEnum = true
			}
		}
		if !isValidEnum {
			return val, gqlerror.ErrorPathf(v.path, "%s is not a valid %s", val.String(), def.Name)
		}
		return val, nil
	case ast.Scalar:
		kind := val.Type().Kind()
		switch typ.NamedType {
		case "Int", "Int64":
			if kind == reflect.String || kind == reflect.Int || kind == reflect.Int32 || kind == reflect.Int64 {
				var errIntCoerce error
				var valString string
				if kind == reflect.String {
					valString = val.String()
				} else {
					valString = strconv.FormatInt(val.Int(), 10)
				}
				if typ.NamedType == "Int" {
					_, errIntCoerce = strconv.ParseInt(valString, 10, 32)
				} else {
					_, errIntCoerce = strconv.ParseInt(valString, 10, 64)
				}
				if errIntCoerce != nil {
					if errors.Is(errIntCoerce, strconv.ErrRange) {
						return val, gqlerror.ErrorPathf(v.path, "Out of range value '%s', for type `%s`", valString, typ.NamedType)

					} else {
						return val, gqlerror.ErrorPathf(v.path, "Type mismatched for Value `%s`, expected:`%s`", valString, typ.NamedType)
					}
				}
				return val, nil
			}

		case "Float":
			if kind == reflect.String || kind == reflect.Float32 || kind == reflect.Float64 || kind == reflect.Int || kind == reflect.Int32 || kind == reflect.Int64 {
				return val, nil
			}
		case "String":
			if kind == reflect.String {
				return val, nil
			}

		case "Boolean":
			if kind == reflect.Bool {
				return val, nil
			}

		case "ID":
			if kind == reflect.Int || kind == reflect.Int32 || kind == reflect.Int64 || kind == reflect.String {
				if val.Type().Name() != "string" {
					val = val.Convert((reflect.ValueOf("string")).Type())
				}
				return val, nil
			}
		default:
			// assume custom scalars are ok
			return val, nil
		}
		return val, gqlerror.ErrorPathf(v.path, "cannot use %s as %s", kind.String(), typ.NamedType)
	case ast.InputObject:
		if val.Kind() != reflect.Map {
			return val, gqlerror.ErrorPathf(v.path, "must be a %s", def.Name)
		}

		// check for unknown fields
		for _, name := range val.MapKeys() {
			val.MapIndex(name)
			fieldDef := def.Fields.ForName(name.String())
			resetPath()
			v.path = append(v.path, ast.PathName(name.String()))

			if fieldDef == nil {
				return val, gqlerror.ErrorPathf(v.path, "unknown field")
			}
		}

		for _, fieldDef := range def.Fields {
			resetPath()
			v.path = append(v.path, ast.PathName(fieldDef.Name))

			field := val.MapIndex(reflect.ValueOf(fieldDef.Name))
			if !field.IsValid() {
				if fieldDef.Type.NonNull {
					if fieldDef.DefaultValue != nil {
						var err error
						_, err = fieldDef.DefaultValue.Value(nil)
						if err == nil {
							continue
						}
					}
					return val, gqlerror.ErrorPathf(v.path, "must be defined")
				}
				continue
			}

			if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
				if fieldDef.Type.NonNull && field.IsNil() {
					return val, gqlerror.ErrorPathf(v.path, "cannot be null")
				}
				//allow null object field and skip it
				if !fieldDef.Type.NonNull && field.IsNil() {
					continue
				}
				field = field.Elem()
			}
			cval, err := v.validateVarType(fieldDef.Type, field)
			if err != nil {
				return val, err
			}
			val.SetMapIndex(reflect.ValueOf(fieldDef.Name), cval)
		}
	default:
		panic(fmt.Errorf("unsupported type %s", def.Kind))
	}
	return val, nil
}
