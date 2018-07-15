package validator

import (
	"github.com/vektah/gqlparser/ast"
	. "github.com/vektah/gqlparser/validator"
)

func init() {
	AddRule("VariablesInAllowedPosition", func(observers *Events, addError AddErrFunc) {
		var varDefs ast.VariableDefinitions

		observers.OnOperation(func(walker *Walker, operation *ast.OperationDefinition) {
			varDefs = operation.VariableDefinitions
		})

		observers.OnOperationLeave(func(walker *Walker, operation *ast.OperationDefinition) {
			varDefs = nil
		})

		observers.OnValue(func(walker *Walker, value *ast.Value) {
			if value.Kind != ast.Variable || value.ExpectedType == nil {
				return
			}

			varDef := varDefs.Find(value.Raw)
			if varDef == nil {
				return
			}

			// If there is a default non nullable types can be null
			if varDef.DefaultValue != nil && varDef.DefaultValue.Kind != ast.NullValue {
				notNull, isNotNull := value.ExpectedType.(ast.NonNullType)
				if isNotNull {
					value.ExpectedType = notNull.Type
				}
			}

			if !varDef.Type.IsCompatible(value.ExpectedType) {
				addError(
					Message(
						`Variable "$%s" of type "%s" used in position expecting type "%s".`,
						value,
						varDef.Type.String(),
						value.ExpectedType.String(),
					),
				)
			}
		})
	})
}
