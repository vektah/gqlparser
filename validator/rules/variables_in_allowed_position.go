package validator

import (
	"github.com/vektah/gqlparser/ast"
	. "github.com/vektah/gqlparser/validator"
)

func init() {
	AddRule("VariablesInAllowedPosition", func(observers *Events, addError AddErrFunc) {
		observers.OnValue(func(walker *Walker, value *ast.Value) {
			if value.Kind != ast.Variable || value.ExpectedType == nil || value.VariableDefinition == nil || walker.CurrentOperation == nil {
				return
			}

			// todo: move me into walk
			// If there is a default non nullable types can be null
			if value.VariableDefinition.DefaultValue != nil && value.VariableDefinition.DefaultValue.Kind != ast.NullValue {
				notNull, isNotNull := value.ExpectedType.(ast.NonNullType)
				if isNotNull {
					value.ExpectedType = notNull.Type
				}
			}

			if !value.VariableDefinition.Type.IsCompatible(value.ExpectedType) {
				addError(
					Message(
						`Variable "$%s" of type "%s" used in position expecting type "%s".`,
						value,
						value.VariableDefinition.Type.String(),
						value.ExpectedType.String(),
					),
				)
			}
		})
	})
}
