package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	addRule("VariablesAreInputTypes", func(observers *Events, addError addErrFunc) {
		observers.OnVariable(func(walker *Walker, valueType gqlparser.Type, def *gqlparser.Definition, variable gqlparser.VariableDefinition) {
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
