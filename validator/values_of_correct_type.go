package validator

import (
	"fmt"

	"github.com/vektah/gqlparser"
)

func init() {
	addRule("ValuesOfCorrectType", func(observers *Events, addError addErrFunc) {
		observers.OnValue(func(walker *Walker, expectedType gqlparser.Type, def *gqlparser.Definition, value gqlparser.Value) {
			if def == nil || expectedType == nil {
				return
			}

			if def.Kind == gqlparser.Scalar {
				// Skip custom validating scalars
				if !def.OneOf("Int", "Float", "String", "Boolean", "ID") {
					return
				}
			}

			validateValue(expectedType, def, value, addError)
		})
	})
}

func validateValue(expectedType gqlparser.Type, def *gqlparser.Definition, value gqlparser.Value, addError addErrFunc) {
	var possibleEnums []string
	if def.Kind == gqlparser.Enum {
		for _, val := range def.Values {
			possibleEnums = append(possibleEnums, val.Name)
		}
	}

	rawVal, err := value.Value(nil)
	if err != nil {
		unexpectedTypeMessage(addError, expectedType.String(), value.String())
	}

	switch value := value.(type) {
	case gqlparser.NullValue:
		if _, nonNullable := expectedType.(gqlparser.NonNullType); nonNullable {
			unexpectedTypeMessage(addError, expectedType.String(), value.String())
		}

	case gqlparser.ListValue:
		listType, isList := expectedType.(gqlparser.ListType)
		if !isList {
			unexpectedTypeMessage(addError, expectedType.String(), value.String())
		}

		for _, item := range value {
			validateValue(listType.Type, def, item, addError)
		}

	case gqlparser.IntValue:
		if !def.OneOf("Int", "Float", "ID") {
			unexpectedTypeMessage(addError, expectedType.String(), value.String())
		}

	case gqlparser.FloatValue:
		if !def.OneOf("Float") {
			unexpectedTypeMessage(addError, expectedType.String(), value.String())
		}

	case gqlparser.StringValue, gqlparser.BlockValue:
		if def.Kind == gqlparser.Enum {
			rawValStr := fmt.Sprint(rawVal)
			addError(
				Message("Expected type %s, found %s.", expectedType.String(), value.String()),
				SuggestListUnquoted("Did you mean the enum value", rawValStr, possibleEnums),
			)
		} else if !def.OneOf("String", "ID") {
			unexpectedTypeMessage(addError, expectedType.String(), value.String())
		}

	case gqlparser.EnumValue:
		if def.Kind != gqlparser.Enum || def.EnumValue(string(value)) == nil {
			rawValStr := fmt.Sprint(rawVal)
			addError(
				Message("Expected type %s, found %s.", expectedType.String(), value.String()),
				SuggestListUnquoted("Did you mean the enum value", rawValStr, possibleEnums),
			)
		}

	case gqlparser.BooleanValue:
		if !def.OneOf("Boolean") {
			unexpectedTypeMessage(addError, expectedType.String(), value.String())
		}

	case gqlparser.ObjectValue:

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

	case gqlparser.Variable:
		return

	default:
		panic(fmt.Errorf("unhandled %T", value))
	}
}

func unexpectedTypeMessage(addError addErrFunc, expected, value string) {
	addError(Message("Expected type %s, found %s.", expected, value))
}
