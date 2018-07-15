package validator

import (
	"strconv"

	"github.com/vektah/gqlparser/ast"
)

func init() {
	addRule("SingleFieldSubscriptions", func(observers *Events, addError addErrFunc) {
		observers.OnOperation(func(walker *Walker, operation *ast.OperationDefinition) {
			if operation.Operation != ast.Subscription {
				return
			}

			if len(operation.SelectionSet) != 1 {
				name := "Anonymous Subscription"
				if operation.Name != "" {
					name = `Subscription ` + strconv.Quote(operation.Name)
				}

				addError(
					Message(`%s must select only one top level field.`, name),
				)
			}
		})
	})
}
