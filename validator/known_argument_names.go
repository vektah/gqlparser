package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	fieldVisitors = append(fieldVisitors, knownFieldArgumentNames)
	directiveVisitors = append(directiveVisitors, knownDirectiveArgumentNames)
}

// A GraphQL field is only valid if all supplied arguments are defined by that field.
func knownFieldArgumentNames(ctx *vctx, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field) {
	if fieldDef == nil {
		return
	}
	for _, arg := range field.Arguments {
		def := fieldDef.Arguments.ForName(arg.Name)
		if def != nil {
			continue
		}

		var suggestions []string
		for _, argDef := range fieldDef.Arguments {
			suggestions = append(suggestions, argDef.Name)
		}

		ctx.errors = append(ctx.errors, Error(
			Rule("KnownArgumentNames"),
			Message(`Unknown argument "%s" on field "%s" of type "%s".`, arg.Name, field.Name, parentDef.Name),
			SuggestList(arg.Name, suggestions),
		))
	}
}

func knownDirectiveArgumentNames(ctx *vctx, parentDef *gqlparser.Definition, directiveDef *gqlparser.DirectiveDefinition, directive *gqlparser.Directive, location gqlparser.DirectiveLocation) {
	if directiveDef == nil {
		return
	}
	for _, arg := range directive.Arguments {
		def := directiveDef.Arguments.ForName(arg.Name)
		if def != nil {
			continue
		}

		var suggestions []string
		for _, argDef := range directiveDef.Arguments {
			suggestions = append(suggestions, argDef.Name)
		}

		ctx.errors = append(ctx.errors, Error(
			Rule("KnownArgumentNames"),
			Message(`Unknown argument "%s" on directive "@%s".`, arg.Name, directive.Name),
			SuggestList(arg.Name, suggestions),
		))
	}
}
