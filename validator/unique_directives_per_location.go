package validator

import (
	"github.com/vektah/gqlparser/ast"
)

func init() {
	addRule("UniqueDirectivesPerLocation", func(observers *Events, addError addErrFunc) {
		observers.OnDirectiveList(func(walker *Walker, parentDef *ast.Definition, directives []ast.Directive, location ast.DirectiveLocation) {
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
