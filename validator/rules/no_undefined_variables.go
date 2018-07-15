package validator

import (
	"github.com/vektah/gqlparser/ast"
	. "github.com/vektah/gqlparser/validator"
)

func init() {
	AddRule("NoUndefinedVariables", func(observers *Events, addError AddErrFunc) {
		observers.OnValue(func(walker *Walker, value *ast.Value) {
			if walker.CurrentOperation == nil || value.Kind != ast.Variable || value.VariableDefinition != nil {
				return
			}

			if walker.CurrentOperation.Name != "" {
				addError(Message(`Variable "$%s" is not defined by operation "%s".`, value, walker.CurrentOperation.Name))
			} else {
				addError(Message(`Variable "$%s" is not defined.`, value))
			}
		})
	})
}
