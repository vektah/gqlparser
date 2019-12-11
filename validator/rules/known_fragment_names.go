package validator

import (
	"github.com/dgraph-io/gqlparser/ast"
	. "github.com/dgraph-io/gqlparser/validator"
)

func init() {
	AddRule("KnownFragmentNames", func(observers *Events, addError AddErrFunc) {
		observers.OnFragmentSpread(func(walker *Walker, fragmentSpread *ast.FragmentSpread) {
			if fragmentSpread.Definition == nil {
				addError(
					Message(`Unknown fragment "%s".`, fragmentSpread.Name),
					At(fragmentSpread.Position),
				)
			}
		})
	})
}
