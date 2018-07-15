package validator

import (
	"github.com/vektah/gqlparser/ast"
	. "github.com/vektah/gqlparser/validator"
)

func init() {
	AddRule("NoUndefinedVariables", func(observers *Events, addError AddErrFunc) {

		var currentOperation *ast.OperationDefinition

		observers.OnOperation(func(walker *Walker, operation *ast.OperationDefinition) {
			currentOperation = operation
		})

		observers.OnOperationLeave(func(walker *Walker, operation *ast.OperationDefinition) {
			currentOperation = nil
		})

		observers.OnValue(func(walker *Walker, valueType ast.Type, def *ast.Definition, value *ast.Value) {
			if currentOperation == nil {
				// not in operation context
				return
			}

			var variables []string
			var filterVariable func(value *ast.Value)
			filterVariable = func(value *ast.Value) {
				switch value.Kind {
				case ast.Variable:
					variables = append(variables, value.Raw)
				case ast.ListValue, ast.ObjectValue:
					for _, v := range value.Children {
						filterVariable(v.Value)
					}
				default:
					return
				}
			}
			filterVariable(value)

		variables:
			for _, variable := range variables {
				for _, varDef := range currentOperation.VariableDefinitions {
					if varDef.Variable == variable {
						continue variables
					}
				}
				if currentOperation.Name != "" {
					addError(Message(`Variable "$%s" is not defined by operation "%s".`, value, currentOperation.Name))
				} else {
					addError(Message(`Variable "$%s" is not defined.`, value))
				}
			}
		})
	})
}
