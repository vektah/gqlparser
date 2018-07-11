package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	addRule("UniqueOperationNames", func(observers *Events, addError addErrFunc) {
		seen := map[string]bool{}

		observers.OnOperation(func(walker *Walker, operation *gqlparser.OperationDefinition) {
			if seen[operation.Name] {
				addError(
					Message(`There can be only one operation named "%s".`, operation.Name),
				)
			}
			seen[operation.Name] = true
		})
	})
}
