package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	directiveListVisitors = append(directiveListVisitors, uniqueDirectivesPerLocation)
}

// A GraphQL document is only valid if all directives at a given location are uniquely named.
func uniqueDirectivesPerLocation(ctx *vctx, parentDef *gqlparser.Definition, directives []gqlparser.Directive, location gqlparser.DirectiveLocation) {
	seen := map[string]bool{}

	for _, dir := range directives {
		if seen[dir.Name] {
			ctx.errors = append(ctx.errors, Error(
				Rule("UniqueDirectivesPerLocation"),
				Message(`The directive "%s" can only be used once at this location.`, dir.Name),
			))
		}
		seen[dir.Name] = true
	}
}
