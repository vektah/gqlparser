package validator

import (
	"github.com/vektah/gqlparser/ast"
)

func init() {
	addRule("LoneAnonymousOperation", func(observers *Events, addError addErrFunc) {
		observers.OnOperation(func(walker *Walker, operation *ast.OperationDefinition) {
			if operation.Name == "" && len(walker.Document.Operations) > 1 {
				addError(Message(`This anonymous operation must be the only defined operation.`))
			}
		})
	})
}
