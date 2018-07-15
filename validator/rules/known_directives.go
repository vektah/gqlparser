package validator

import (
	"github.com/vektah/gqlparser/ast"
	. "github.com/vektah/gqlparser/validator"
)

func init() {
	AddRule("KnownDirectives", func(observers *Events, addError AddErrFunc) {
		observers.OnDirective(func(walker *Walker, parentDef *ast.Definition, directiveDef *ast.DirectiveDefinition, directive *ast.Directive, location ast.DirectiveLocation) {
			if directiveDef == nil {
				addError(
					Message(`Unknown directive "%s".`, directive.Name),
				)
				return
			}

			for _, loc := range directiveDef.Locations {
				if loc == location {
					return
				}
			}

			addError(
				Message(`Directive "%s" may not be used on %s.`, directive.Name, location),
			)
		})
	})
}
