package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	addRule("KnownArgumentNames", func(observers *Events, addError addErrFunc) {
		// A GraphQL field is only valid if all supplied arguments are defined by that field.
		observers.OnField(func(walker *Walker, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field) {
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

				addError(
					Message(`Unknown argument "%s" on field "%s" of type "%s".`, arg.Name, field.Name, parentDef.Name),
					SuggestListQuoted("Did you mean", arg.Name, suggestions),
				)
			}
		})

		observers.OnDirective(func(walker *Walker, parentDef *gqlparser.Definition, directiveDef *gqlparser.DirectiveDefinition, directive *gqlparser.Directive, location gqlparser.DirectiveLocation) {
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

				addError(
					Message(`Unknown argument "%s" on directive "@%s".`, arg.Name, directive.Name),
					SuggestListQuoted("Did you mean", arg.Name, suggestions),
				)
			}
		})
	})
}
