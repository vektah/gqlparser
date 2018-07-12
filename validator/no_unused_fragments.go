package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	addRule("NoUnusedFragments", func(observers *Events, addError addErrFunc) {

		inFragmentDefinition := false
		fragmentNameUsed := make(map[string]bool)

		observers.OnFragmentSpread(func(walker *Walker, parentDef *gqlparser.Definition, fragmentDef *gqlparser.FragmentDefinition, fragmentSpread *gqlparser.FragmentSpread) {
			if !inFragmentDefinition {
				fragmentNameUsed[fragmentSpread.Name] = true
			}
		})

		observers.OnFragment(func(walker *Walker, parentDef *gqlparser.Definition, fragment *gqlparser.FragmentDefinition) {
			inFragmentDefinition = true
			if !fragmentNameUsed[fragment.Name] {
				addError(Message(`Fragment "%s" is never used.`, fragment.Name))
			}
		})
	})
}
