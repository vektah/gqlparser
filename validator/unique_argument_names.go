package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	fieldVisitors = append(fieldVisitors, uniqueFieldArgumentNames)
	directiveVisitors = append(directiveVisitors, uniqueDirectiveArgumentNames)
}

func checkUniqueArgs(ctx *vctx, args []gqlparser.Argument) {
	knownArgNames := map[string]bool{}

	for _, arg := range args {
		if knownArgNames[arg.Name] {
			ctx.errors = append(ctx.errors, Error(
				Rule("UniqueArgumentNames"),
				Message(`There can be only one argument named "%s".`, arg.Name),
			))
		}

		knownArgNames[arg.Name] = true
	}
}

// A GraphQL field is only valid if all supplied arguments are defined by that field.
func uniqueFieldArgumentNames(ctx *vctx, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field) {
	checkUniqueArgs(ctx, field.Arguments)
}

func uniqueDirectiveArgumentNames(ctx *vctx, parentDef *gqlparser.Definition, directiveDef *gqlparser.DirectiveDefinition, directive *gqlparser.Directive, location gqlparser.DirectiveLocation) {
	checkUniqueArgs(ctx, directive.Arguments)
}
