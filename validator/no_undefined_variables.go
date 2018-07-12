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

		observers.OnArgument(func(walker *Walker, arg *gqlparser.Argument) {
			if currentOperation == nil {
				// not in operation context
				return
			}
			variable, isVariable := arg.Value.(gqlparser.Variable)
			if !isVariable {
				return
			}
			for _, varDef := range currentOperation.VariableDefinitions {
				if varDef.Variable == variable {
					return
				}
			}
			if currentOperation.Name != "" {
				addError(Message(`Variable "$%s" is not defined by operation "%s".`, arg.Name, currentOperation.Name))
			} else {
				addError(Message(`Variable "$%s" is not defined.`, arg.Name))
			}
		})
	})
}
