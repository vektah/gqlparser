package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	addRule("UniqueArgumentNames", func(observers *Events, addError addErrFunc) {
		observers.OnField(func(walker *Walker, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field) {
			checkUniqueArgs(field.Arguments, addError)
		})

		observers.OnDirective(func(walker *Walker, parentDef *gqlparser.Definition, directiveDef *gqlparser.DirectiveDefinition, directive *gqlparser.Directive, location gqlparser.DirectiveLocation) {
			checkUniqueArgs(directive.Arguments, addError)
		})
	})
}

func checkUniqueArgs(args []gqlparser.Argument, addError addErrFunc) {
	knownArgNames := map[string]bool{}

	for _, arg := range args {
		if knownArgNames[arg.Name] {
			addError(
				Message(`There can be only one argument named "%s".`, arg.Name),
			)
		}

		knownArgNames[arg.Name] = true
	}
}
