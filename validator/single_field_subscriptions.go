package validator

import (
	"strconv"

	"github.com/vektah/gqlparser"
)

func init() {
	addRule("SingleFieldSubscriptions", func(observers *Events, addError addErrFunc) {
		observers.OnOperation(func(walker *Walker, operation *gqlparser.OperationDefinition) {
			if operation.Operation != gqlparser.Subscription {
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
