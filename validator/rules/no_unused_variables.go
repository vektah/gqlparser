package validator

import (
	"github.com/vektah/gqlparser/ast"
	. "github.com/vektah/gqlparser/validator"
)

func init() {
	AddRule("NoUnusedVariables", func(observers *Events, addError AddErrFunc) {

		var variableNameUsed map[string]bool

		observers.OnOperation(func(walker *Walker, operation *ast.OperationDefinition) {
			variableNameUsed = make(map[string]bool)
		})

		observers.OnOperationLeave(func(walker *Walker, operation *ast.OperationDefinition) {
			for _, varDef := range operation.VariableDefinitions {
				if variableNameUsed[string(varDef.Variable)] {
					continue
				}

				if operation.Name != "" {
					addError(Message(`Variable "$%s" is never used in operation "%s".`, varDef.Variable, operation.Name))
				} else {
					addError(Message(`Variable "$%s" is never used.`, varDef.Variable))
				}
			}

			variableNameUsed = nil
		})

		observers.OnValue(func(walker *Walker, value *ast.Value) {
			if variableNameUsed == nil {
				// not in operation context
				return
			}
			if value.Kind != ast.Variable {
				return
			}
			variableNameUsed[value.Raw] = true
		})
	})
}
