package validator

import (
	"github.com/vektah/gqlparser/ast"
)

func init() {
	addRule("VariablesInAllowedPosition", func(observers *Events, addError addErrFunc) {
		var varDefs ast.VariableDefinitions

		observers.OnOperation(func(walker *Walker, operation *ast.OperationDefinition) {
			varDefs = operation.VariableDefinitions
		})

		observers.OnOperationLeave(func(walker *Walker, operation *ast.OperationDefinition) {
			varDefs = nil
		})

		observers.OnValue(func(walker *Walker, expectedType ast.Type, def *ast.Definition, value ast.Value) {
			if def == nil || expectedType == nil || varDefs == nil {
				return
			}

			validateVariable(walker, expectedType, def, value, addError, varDefs)
		})
	})
}

func validateVariable(walker *Walker, expectedType ast.Type, def *ast.Definition, value ast.Value, addError addErrFunc, varDefs ast.VariableDefinitions) {
	switch value := value.(type) {

	case ast.ListValue:
		listType, isList := expectedType.(ast.ListType)
		if !isList {
			return
		}

		for _, item := range value {
			validateVariable(walker, listType.Type, def, item, addError, varDefs)
		}

	case ast.Variable:
		varDef := varDefs.Find(string(value))
		if varDef == nil {
			return
		}

		// If there is a default non nullable types can be null
		if varDef.DefaultValue != nil {
			if _, isNullvalue := varDef.DefaultValue.(ast.NullValue); !isNullvalue {
				notNull, isNotNull := expectedType.(ast.NonNullType)
				if isNotNull {
					expectedType = notNull.Type
				}
			}
		}

		if !varDef.Type.IsCompatible(expectedType) {
			addError(
				Message(
					`Variable "$%s" of type "%s" used in position expecting type "%s".`,
					value,
					varDef.Type.String(),
					expectedType.String(),
				),
			)
		}
	}
}
