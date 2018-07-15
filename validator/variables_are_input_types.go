package validator

import (
	"github.com/vektah/gqlparser/ast"
)

func init() {
	addRule("VariablesAreInputTypes", func(observers *Events, addError addErrFunc) {
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
