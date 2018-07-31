package validator

import (
	"fmt"

	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/coerce"
	. "github.com/vektah/gqlparser/validator"
)

func isEnumIssue(value *ast.Value) bool {
	if (value.Kind == ast.StringValue || value.Kind == ast.BlockValue) && value.Definition.Kind == ast.Enum {
		return true
	}

	if value.Kind == ast.EnumValue && (value.Definition.Kind != ast.Enum || value.Definition.EnumValues.ForName(value.Raw) == nil) {
		return true
	}

	return false
}

func init() {
	AddRule("ValuesOfCorrectType", func(observers *Events, addError AddErrFunc) {
		observers.OnValue(func(walker *Walker, value *ast.Value) {
			if value.Definition == nil || value.ExpectedType == nil {
				return
			}

			if value.Kind == ast.Variable {
				return
			}

			if isEnumIssue(value) {
				var possibleEnums []string
				if value.Definition.Kind == ast.Enum {
					for _, val := range value.Definition.EnumValues {
						possibleEnums = append(possibleEnums, val.Name)
					}
				}

				addError(
					Message("Expected type %s, found %s.", value.ExpectedType.String(), value.String()),
					SuggestListUnquoted("Did you mean the enum value", value.Raw, possibleEnums),
					At(value.Position),
				)
				return
			}

			if err := validateType(walker, value); err != nil {
				if err == coerce.UnexpectedType {
					addError(
						Message("Expected type %s, found %s.", value.ExpectedType.String(), value.String()),
						At(value.Position),
					)
				} else {
					addError(
						Message("Expected type %s, found %s; "+err.Error(), value.ExpectedType.String(), value.String()),
						At(value.Position),
					)
				}
				return
			}

			if value.ExpectedType.NamedType != "" && value.Definition.Kind == ast.InputObject {
				for _, field := range value.Definition.Fields {
					if field.Type.NonNull {
						fieldValue := value.Children.ForName(field.Name)
						if fieldValue == nil && field.DefaultValue == nil {
							addError(
								Message("Field %s.%s of required type %s was not provided.", value.Definition.Name, field.Name, field.Type.String()),
								At(value.Position),
							)
							continue
						}
					}
				}

				for _, fieldValue := range value.Children {
					if value.Definition.Fields.ForName(fieldValue.Name) == nil {
						var suggestions []string
						for _, fieldValue := range value.Definition.Fields {
							suggestions = append(suggestions, fieldValue.Name)
						}

						addError(
							Message(`Field "%s" is not defined by type %s.`, fieldValue.Name, value.Definition.Name),
							SuggestListUnquoted("Did you mean", fieldValue.Name, suggestions),
							At(fieldValue.Position),
						)
					}
				}
				return
			}
		})
	})
}

func validateType(walker *Walker, value *ast.Value) error {
	if value.Kind == ast.Variable {
		return nil
	}

	if value.Kind == ast.NullValue {
		if value.ExpectedType.NonNull {
			return coerce.UnexpectedType
		}
		return nil
	}

	if value.ExpectedType.Elem != nil {
		if value.Kind != ast.ListValue {
			cpy := *value
			cpy.ExpectedType = cpy.ExpectedType.Elem
			return validateType(walker, &cpy)
		}
		return nil
	}

	if value.Definition.Kind == ast.InputObject {
		if value.Kind != ast.ObjectValue {
			return coerce.UnexpectedType
		}
		return nil
	}

	if value.ExpectedType.Elem == nil && value.Definition.IsLeafType() {
		goVal, err := value.Value(nil)
		if err != nil {
			return coerce.UnexpectedType
		}
		_, err = walker.CoerceScalar(value.ExpectedType, value.Definition, goVal)
		return err
	}

	return fmt.Errorf("Unexpected type passed to isTypeValid %s expected %s", value.String(), value.ExpectedType.String())
}
