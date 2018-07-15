package validator

import (
	"fmt"

	"github.com/vektah/gqlparser/ast"
)

func init() {
	addRule("ValuesOfCorrectType", func(observers *Events, addError addErrFunc) {
		observers.OnValue(func(walker *Walker, expectedType ast.Type, def *ast.Definition, value ast.Value) {
			if def == nil || expectedType == nil {
				return
			}

			if def.Kind == ast.Scalar {
				// Skip custom validating scalars
				if !def.OneOf("Int", "Float", "String", "Boolean", "ID") {
					return
				}
			}

			validateValue(expectedType, def, value, addError)
		})
	})
}

func validateValue(expectedType ast.Type, def *ast.Definition, value ast.Value, addError addErrFunc) {
	var possibleEnums []string
	if def.Kind == ast.Enum {
		for _, val := range def.Values {
			possibleEnums = append(possibleEnums, val.Name)
		}
	}

	rawVal, err := value.Value(nil)
	if err != nil {
		unexpectedTypeMessage(addError, expectedType.String(), value.String())
	}

	switch value := value.(type) {
	case ast.NullValue:
		if _, nonNullable := expectedType.(ast.NonNullType); nonNullable {
			unexpectedTypeMessage(addError, expectedType.String(), value.String())
		}

	case ast.ListValue:
		listType, isList := expectedType.(ast.ListType)
		if !isList {
			unexpectedTypeMessage(addError, expectedType.String(), value.String())
		}

		for _, item := range value {
			validateValue(listType.Type, def, item, addError)
		}

	case ast.IntValue:
		if !def.OneOf("Int", "Float", "ID") {
			unexpectedTypeMessage(addError, expectedType.String(), value.String())
		}

	case ast.FloatValue:
		if !def.OneOf("Float") {
			unexpectedTypeMessage(addError, expectedType.String(), value.String())
		}

	case ast.StringValue, ast.BlockValue:
		if def.Kind == ast.Enum {
			rawValStr := fmt.Sprint(rawVal)
			addError(
				Message("Expected type %s, found %s.", expectedType.String(), value.String()),
				SuggestListUnquoted("Did you mean the enum value", rawValStr, possibleEnums),
			)
		} else if !def.OneOf("String", "ID") {
			unexpectedTypeMessage(addError, expectedType.String(), value.String())
		}

	case ast.EnumValue:
		if def.Kind != ast.Enum || def.EnumValue(string(value)) == nil {
			rawValStr := fmt.Sprint(rawVal)
			addError(
				Message("Expected type %s, found %s.", expectedType.String(), value.String()),
				SuggestListUnquoted("Did you mean the enum value", rawValStr, possibleEnums),
			)
		}

	case ast.BooleanValue:
		if !def.OneOf("Boolean") {
			unexpectedTypeMessage(addError, expectedType.String(), value.String())
		}

	case ast.ObjectValue:

		for _, field := range def.Fields {
			if field.Type.IsRequired() {
				fieldValue := value.Find(field.Name)
				if fieldValue == nil && field.DefaultValue == nil {
					addError(
						Message("Field %s.%s of required type %s was not provided.", def.Name, field.Name, field.Type.String()),
					)
					continue
				}
			}
		}

		for _, fieldValue := range value {
			if def.Field(fieldValue.Name) == nil {
				var suggestions []string
				for _, fieldValue := range def.Fields {
					suggestions = append(suggestions, fieldValue.Name)
				}

				addError(
					Message(`Field "%s" is not defined by type %s.`, fieldValue.Name, def.Name),
					SuggestListUnquoted("Did you mean", fieldValue.Name, suggestions),
				)
			}
		}

	case ast.Variable:
		return

	default:
		panic(fmt.Errorf("unhandled %T", value))
	}
}

func unexpectedTypeMessage(addError addErrFunc, expected, value string) {
	addError(Message("Expected type %s, found %s.", expected, value))
}
