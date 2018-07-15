package validator

import (
	"fmt"

	"github.com/vektah/gqlparser/ast"
	. "github.com/vektah/gqlparser/validator"
)

func init() {
	AddRule("KnownFragmentNames", func(observers *Events, addError AddErrFunc) {
		observers.OnFragmentSpread(func(walker *Walker, parentDef *ast.Definition, fragmentDef *ast.FragmentDefinition, fragmentSpread *ast.FragmentSpread) {
			if fragmentDef != nil {
				return
			}

			message := fmt.Sprintf(`Unknown fragment "%s".`, fragmentSpread.Name)
			addError(Message(message))
		})
	})
}
