package validator

import "github.com/vektah/gqlparser"

func init() {
	addRule("NoUndefinedVariables", func(observers *Events, addError addErrFunc) {

		var currentOperation *gqlparser.OperationDefinition

		observers.OnOperation(func(walker *Walker, operation *gqlparser.OperationDefinition) {
			currentOperation = operation
		})

		observers.OnOperationLeave(func(walker *Walker, operation *gqlparser.OperationDefinition) {
			currentOperation = nil
		})

		observers.OnValue(func(walker *Walker, valueType gqlparser.Type, def *gqlparser.Definition, value gqlparser.Value) {
			if currentOperation == nil {
				// not in operation context
				return
			}

			var variables []gqlparser.Variable
			var filterVariable func(value gqlparser.Value)
			filterVariable = func(value gqlparser.Value) {
				switch value := value.(type) {
				case gqlparser.Variable:
					variables = append(variables, value)
				case gqlparser.ListValue:
					for _, v := range value {
						filterVariable(v)
					}
				case gqlparser.ObjectValue:
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
