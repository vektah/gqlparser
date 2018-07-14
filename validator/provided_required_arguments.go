package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	addRule("ProvidedRequiredArguments", func(observers *Events, addError addErrFunc) {

		observers.OnField(func(walker *Walker, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field) {
			if fieldDef == nil {
				return
			}

		argDef:
			for _, argDef := range fieldDef.Arguments {
				if !argDef.Type.IsRequired() {
					continue
				}
				if argDef.DefaultValue != nil {
					continue
				}
				for _, arg := range field.Arguments {
					if arg.Name == argDef.Name {
						continue argDef
					}
				}

				addError(Message(`Field "%s" argument "%s" of type "%s" is required but not provided.`, field.Name, argDef.Name, argDef.Type.String()))
			}
		})

		observers.OnDirective(func(walker *Walker, parentDef *gqlparser.Definition, directiveDef *gqlparser.DirectiveDefinition, directive *gqlparser.Directive, location gqlparser.DirectiveLocation) {
			if directiveDef == nil {
				return
			}

		argDef:
			for _, argDef := range directiveDef.Arguments {
				if !argDef.Type.IsRequired() {
					continue
				}
				if argDef.DefaultValue != nil {
					continue
				}
				for _, arg := range directive.Arguments {
					if arg.Name == argDef.Name {
						continue argDef
					}
				}

				addError(Message(`Directive "@%s" argument "%s" of type "%s" is required but not provided.`, directiveDef.Name, argDef.Name, argDef.Type.String()))
			}
		})
	})
}
