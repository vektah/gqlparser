package validator

import (
	"github.com/vektah/gqlparser/ast"
)

func init() {
	addRule("UniqueFragmentNames", func(observers *Events, addError addErrFunc) {
		seenFragments := map[string]bool{}

		observers.OnFragment(func(walker *Walker, parentDef *ast.Definition, fragment *ast.FragmentDefinition) {
			if seenFragments[fragment.Name] {
				addError(
					Message(`There can be only one fragment named "%s".`, fragment.Name),
				)
			}
			seenFragments[fragment.Name] = true
		})
	})
}
