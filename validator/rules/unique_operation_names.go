package validator

import (
	"github.com/pjmd89/gqlparser/v2/ast"
	. "github.com/pjmd89/gqlparser/v2/validator"
)

func init() {
	AddRule("UniqueOperationNames", func(observers *Events, addError AddErrFunc) {
		seen := map[string]bool{}

		observers.OnOperation(func(walker *Walker, operation *ast.OperationDefinition) {
			if seen[operation.Name] {
				addError(
					Message(`There can be only one operation named "%s".`, operation.Name),
					At(operation.Position),
				)
			}
			seen[operation.Name] = true
		})
	})
}
