package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	addRule("UniqueVariableNames", func(observers *Events, addError addErrFunc) {
		observers.OnOperation(func(walker *Walker, operation *gqlparser.OperationDefinition) {
			seen := map[gqlparser.Variable]bool{}
			for _, def := range operation.VariableDefinitions {
				if seen[def.Variable] {
					addError(Message(`There can be only one variable named "%s".`, def.Variable))
				}
				seen[def.Variable] = true
			}
		})
	})
}
