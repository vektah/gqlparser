package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	addRule("KnownDirectives", func(observers *Events, addError addErrFunc) {
		observers.OnDirective(func(walker *Walker, parentDef *gqlparser.Definition, directiveDef *gqlparser.DirectiveDefinition, directive *gqlparser.Directive, location gqlparser.DirectiveLocation) {
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
