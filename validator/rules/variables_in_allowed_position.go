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

		observers.OnValue(func(walker *Walker, expectedType ast.Type, def *ast.Definition, value *ast.Value) {
			if def == nil || expectedType == nil || varDefs == nil {
				return
			}

			validateVariable(walker, expectedType, def, value, addError, varDefs)
		})
	})
}

func validateVariable(walker *Walker, expectedType ast.Type, def *ast.Definition, value *ast.Value, addError AddErrFunc, varDefs ast.VariableDefinitions) {
	switch value.Kind {
	case ast.ListValue:
		listType, isList := expectedType.(ast.ListType)
		if !isList {
			return
		}

		for _, item := range value.Children {
			validateVariable(walker, listType.Type, def, item.Value, addError, varDefs)
		}

	case ast.Variable:
		varDef := varDefs.Find(value.Raw)
		if varDef == nil {
			return
		}

		// If there is a default non nullable types can be null
		if varDef.DefaultValue != nil && varDef.DefaultValue.Kind != ast.NullValue {
			notNull, isNotNull := expectedType.(ast.NonNullType)
			if isNotNull {
				expectedType = notNull.Type
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
