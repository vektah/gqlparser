package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	addRule("UniqueDirectivesPerLocation", func(observers *Events, addError addErrFunc) {
		observers.OnDirectiveList(func(walker *Walker, parentDef *gqlparser.Definition, directives []gqlparser.Directive, location gqlparser.DirectiveLocation) {
			seen := map[string]bool{}

			for _, dir := range directives {
				if seen[dir.Name] {
					addError(
						Message(`The directive "%s" can only be used once at this location.`, dir.Name),
					)
				}
				seen[dir.Name] = true
			}
		})
	})
}
