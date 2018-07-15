package validator

import (
	"github.com/vektah/gqlparser/ast"
	. "github.com/vektah/gqlparser/validator"
)

func init() {
	AddRule("UniqueInputFieldNames", func(observers *Events, addError AddErrFunc) {
		observers.OnValue(func(walker *Walker, fieldType ast.Type, def *ast.Definition, value *ast.Value) {
			if value.Kind != ast.ObjectValue {
				return
			}

			seen := map[string]bool{}
			for _, field := range value.Children {
				if seen[field.Name] {
					addError(
						Message(`There can be only one input field named "%s".`, field.Name),
					)
				}
				seen[field.Name] = true
			}
		})
	})
}
