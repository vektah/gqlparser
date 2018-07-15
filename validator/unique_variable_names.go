package validator

import (
	"github.com/vektah/gqlparser/ast"
)

func init() {
	addRule("UniqueVariableNames", func(observers *Events, addError addErrFunc) {
		observers.OnOperation(func(walker *Walker, operation *ast.OperationDefinition) {
			seen := map[ast.Variable]bool{}
			for _, def := range operation.VariableDefinitions {
				if seen[def.Variable] {
					addError(Message(`There can be only one variable named "%s".`, def.Variable))
				}
				seen[def.Variable] = true
			}
		})
	})
}
