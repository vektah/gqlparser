package validator

import (
	"github.com/vektah/gqlparser/ast"
	. "github.com/vektah/gqlparser/validator"
)

func init() {
	AddRule("ProvidedRequiredArguments", func(observers *Events, addError AddErrFunc) {

		observers.OnField(func(walker *Walker, field *ast.Field) {
			if field.Definition == nil {
				return
			}

		argDef:
			for _, argDef := range field.Definition.Arguments {
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

		observers.OnDirective(func(walker *Walker, parentDef *ast.Definition, directiveDef *ast.DirectiveDefinition, directive *ast.Directive, location ast.DirectiveLocation) {
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
