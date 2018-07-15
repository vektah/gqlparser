package validator

import (
	"fmt"

	"github.com/vektah/gqlparser/ast"
)

func init() {
	addRule("KnownFragmentNames", func(observers *Events, addError addErrFunc) {
		observers.OnFragmentSpread(func(walker *Walker, parentDef *ast.Definition, fragmentDef *ast.FragmentDefinition, fragmentSpread *ast.FragmentSpread) {
			if fragmentDef != nil {
				return
			}

			message := fmt.Sprintf(`Unknown fragment "%s".`, fragmentSpread.Name)
			addError(Message(message))
		})
	})
}
