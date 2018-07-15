package validator

import "github.com/vektah/gqlparser/ast"

func init() {
	addRule("NoUndefinedVariables", func(observers *Events, addError addErrFunc) {

		var currentOperation *ast.OperationDefinition

		observers.OnOperation(func(walker *Walker, operation *ast.OperationDefinition) {
			currentOperation = operation
		})

		observers.OnOperationLeave(func(walker *Walker, operation *ast.OperationDefinition) {
			currentOperation = nil
		})

		observers.OnValue(func(walker *Walker, valueType ast.Type, def *ast.Definition, value ast.Value) {
			if currentOperation == nil {
				// not in operation context
				return
			}

			var variables []ast.Variable
			var filterVariable func(value ast.Value)
			filterVariable = func(value ast.Value) {
				switch value := value.(type) {
				case ast.Variable:
					variables = append(variables, value)
				case ast.ListValue:
					for _, v := range value {
						filterVariable(v)
					}
				case ast.ObjectValue:
					for _, v := range value {
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
