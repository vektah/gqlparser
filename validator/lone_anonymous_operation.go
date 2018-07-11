package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	addRule("LoneAnonymousOperation", func(observers *Events, addError addErrFunc) {
		observers.OnOperation(func(walker *Walker, operation *gqlparser.OperationDefinition) {
			if operation.Name == "" && observers.operationCount > 1 {
				addError(Message(`This anonymous operation must be the only defined operation.`))
			}
		})
	})
}
