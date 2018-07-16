package validator

import (
	"fmt"

	"github.com/vektah/gqlparser/ast"
	. "github.com/vektah/gqlparser/validator"
)

func init() {
	AddRule("ValuesOfCorrectType", func(observers *Events, addError AddErrFunc) {
		observers.OnValue(func(walker *Walker, value *ast.Value) {
			if value.Definition == nil || value.ExpectedType == nil {
				return
			}

			if value.Definition.Kind == ast.Scalar {
				// Skip custom validating scalars
				if !value.Definition.OneOf("Int", "Float", "String", "Boolean", "ID") {
					return
				}
			}

			var possibleEnums []string
			if value.Definition.Kind == ast.Enum {
				for _, val := range value.Definition.Values {
					possibleEnums = append(possibleEnums, val.Name)
				}
			}

			rawVal, err := value.Value(nil)
			if err != nil {
				unexpectedTypeMessage(addError, value.ExpectedType.String(), value.String())
			}

			switch value.Kind {
			case ast.NullValue:
				if value.ExpectedType.NonNull {
					unexpectedTypeMessage(addError, value.ExpectedType.String(), value.String())
				}

			case ast.ListValue:
				if value.ExpectedType.Elem == nil {
					unexpectedTypeMessage(addError, value.ExpectedType.String(), value.String())
					return
				}

			case ast.IntValue:
				if !value.Definition.OneOf("Int", "Float", "ID") {
					unexpectedTypeMessage(addError, value.ExpectedType.String(), value.String())
				}

			case ast.FloatValue:
				if !value.Definition.OneOf("Float") {
					unexpectedTypeMessage(addError, value.ExpectedType.String(), value.String())
				}

			case ast.StringValue, ast.BlockValue:
				if value.Definition.Kind == ast.Enum {
					rawValStr := fmt.Sprint(rawVal)
					addError(
						Message("Expected type %s, found %s.", value.ExpectedType.String(), value.String()),
						SuggestListUnquoted("Did you mean the enum value", rawValStr, possibleEnums),
					)
				} else if !value.Definition.OneOf("String", "ID") {
					unexpectedTypeMessage(addError, value.ExpectedType.String(), value.String())
				}

			case ast.EnumValue:
				if value.Definition.Kind != ast.Enum || value.Definition.EnumValue(value.Raw) == nil {
					rawValStr := fmt.Sprint(rawVal)
					addError(
						Message("Expected type %s, found %s.", value.ExpectedType.String(), value.String()),
						SuggestListUnquoted("Did you mean the enum value", rawValStr, possibleEnums),
					)
				}

			case ast.BooleanValue:
				if !value.Definition.OneOf("Boolean") {
					unexpectedTypeMessage(addError, value.ExpectedType.String(), value.String())
				}

			case ast.ObjectValue:

				for _, field := range value.Definition.Fields {
					if field.Type.NonNull {
						fieldValue := value.FieldByName(field.Name)
						if fieldValue == nil && field.DefaultValue == nil {
							addError(
								Message("Field %s.%s of required type %s was not provided.", value.Definition.Name, field.Name, field.Type.String()),
							)
							continue
						}
					}
				}

				for _, fieldValue := range value.Children {
					if value.Definition.Field(fieldValue.Name) == nil {
						var suggestions []string
						for _, fieldValue := range value.Definition.Fields {
							suggestions = append(suggestions, fieldValue.Name)
						}

						addError(
							Message(`Field "%s" is not defined by type %s.`, fieldValue.Name, value.Definition.Name),
							SuggestListUnquoted("Did you mean", fieldValue.Name, suggestions),
						)
					}
				}

			case ast.Variable:
				return

			default:
				panic(fmt.Errorf("unhandled %T", value))
			}
		})
	})
}

func unexpectedTypeMessage(addError AddErrFunc, expected, value string) {
	addError(Message("Expected type %s, found %s.", expected, value))
}
