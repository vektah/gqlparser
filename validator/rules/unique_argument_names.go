package validator

import (
	"github.com/vektah/gqlparser/ast"
	. "github.com/vektah/gqlparser/validator"
)

func init() {
	AddRule("UniqueArgumentNames", func(observers *Events, addError AddErrFunc) {
		observers.OnField(func(walker *Walker, parentDef *ast.Definition, fieldDef *ast.FieldDefinition, field *ast.Field) {
			checkUniqueArgs(field.Arguments, addError)
		})

		observers.OnDirective(func(walker *Walker, parentDef *ast.Definition, directiveDef *ast.DirectiveDefinition, directive *ast.Directive, location ast.DirectiveLocation) {
			checkUniqueArgs(directive.Arguments, addError)
		})
	})
}

func checkUniqueArgs(args []ast.Argument, addError AddErrFunc) {
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
