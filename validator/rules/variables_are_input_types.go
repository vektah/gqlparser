package validator

import (
	"github.com/vektah/gqlparser/ast"
	. "github.com/vektah/gqlparser/validator"
)

func init() {
	AddRule("VariablesAreInputTypes", func(observers *Events, addError AddErrFunc) {
		observers.OnVariable(func(walker *Walker, valueType ast.Type, def *ast.Definition, variable ast.VariableDefinition) {
			if def == nil {
				return
			}
			if !def.IsInputType() {
				addError(
					Message(
						`Variable "$%s" cannot be non-input type "%s".`,
						variable.Variable.String(),
						valueType.String(),
					),
				)
			}
		})
	})
}
