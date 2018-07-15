package validator

import (
	"github.com/vektah/gqlparser/ast"
)

func init() {
	addRule("NoUnusedFragments", func(observers *Events, addError addErrFunc) {

		inFragmentDefinition := false
		fragmentNameUsed := make(map[string]bool)

		observers.OnFragmentSpread(func(walker *Walker, parentDef *ast.Definition, fragmentDef *ast.FragmentDefinition, fragmentSpread *ast.FragmentSpread) {
			if !inFragmentDefinition {
				fragmentNameUsed[fragmentSpread.Name] = true
			}
		})

		observers.OnFragment(func(walker *Walker, parentDef *ast.Definition, fragment *ast.FragmentDefinition) {
			inFragmentDefinition = true
			if !fragmentNameUsed[fragment.Name] {
				addError(Message(`Fragment "%s" is never used.`, fragment.Name))
			}
		})
	})
}
