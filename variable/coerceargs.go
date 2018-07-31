package variable

import (
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/gqlerror"
)

// (objectType, field, variableValues)

func (b *Bag) CoerceArguments(field *ast.Field, variables map[string]interface{}) *gqlerror.Error {
	coercedValues := map[string]interface{}{}

	for _, arg := range field.Definition.Arguments {
		argumentValue := field.Arguments.ForName(arg.Name)
		var hasValue bool
		var value interface{}
		var err error
		if argumentValue != nil {
			if argumentValue.Value.Kind == ast.Variable {
				value, hasValue = b.coercedVars[argumentValue.Value.Raw]
			} else {
				value, err = b.Value(argumentValue.Value)
				if err != nil {
					return gqlerror.ErrorPosf(argumentValue.Position, err.Error())
				}
				hasValue = true
			}
		}

		if !hasValue && arg.DefaultValue != nil {
			value, err = b.Value(arg.DefaultValue)
			hasValue = true
			if err != nil {
				return gqlerror.ErrorPosf(field.Position, err.Error())
			}
		} else if arg.Type.NonNull && (!hasValue || value == nil) {
			return gqlerror.ErrorPosf(field.Position, "argument %s must be provided", arg.Name)
		}

		b.CoerceValue()
		if hasValue {
			if value == nil {

			}
		}
	}
}
