package validator

import "github.com/vektah/gqlparser"

func init() {
	directiveVisitors = append(directiveVisitors, knownDirectives)
}

func knownDirectives(ctx *vctx, parentDef *gqlparser.Definition, directiveDef *gqlparser.DirectiveDefinition, directive *gqlparser.Directive, location gqlparser.DirectiveLocation) {
	if directiveDef == nil {
		ctx.errors = append(ctx.errors, Error(
			Rule("KnownDirectives"),
			Message(`Unknown directive "%s".`, directive.Name),
		))
		return
	}

	for _, loc := range directiveDef.Locations {
		if loc == location {
			return
		}
	}

	ctx.errors = append(ctx.errors, Error(
		Rule("KnownDirectives"),
		Message(`Directive "%s" may not be used on %s.`, directive.Name, location),
	))
}
