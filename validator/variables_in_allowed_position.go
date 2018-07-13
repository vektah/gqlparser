package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	addRule("VariablesInAllowedPosition", func(observers *Events, addError addErrFunc) {
		var varDefs gqlparser.VariableDefinitions

		observers.OnOperation(func(walker *Walker, operation *gqlparser.OperationDefinition) {
			varDefs = operation.VariableDefinitions
		})

		observers.OnOperationLeave(func(walker *Walker, operation *gqlparser.OperationDefinition) {
			varDefs = nil
		})

		observers.OnValue(func(walker *Walker, expectedType gqlparser.Type, def *gqlparser.Definition, value gqlparser.Value) {
			if def == nil || expectedType == nil || varDefs == nil {
				return
			}

			validateVariable(walker, expectedType, def, value, addError, varDefs)
		})
	})
}

func validateVariable(walker *Walker, expectedType gqlparser.Type, def *gqlparser.Definition, value gqlparser.Value, addError addErrFunc, varDefs gqlparser.VariableDefinitions) {
	switch value := value.(type) {

	case gqlparser.ListValue:
		listType, isList := expectedType.(gqlparser.ListType)
		if !isList {
			return
		}

		for _, item := range value {
			validateVariable(walker, listType.Type, def, item, addError, varDefs)
		}

	case gqlparser.Variable:
		varDef := varDefs.Find(string(value))
		if varDef == nil {
			return
		}

		// If there is a default non nullable types can be null
		if varDef.DefaultValue != nil {
			if _, isNullvalue := varDef.DefaultValue.(gqlparser.NullValue); !isNullvalue {
				notNull, isNotNull := expectedType.(gqlparser.NonNullType)
				if isNotNull {
					expectedType = notNull.Type
				}
			}
		}

		if !varDef.Type.IsCompatible(expectedType) {
			addError(
				Message(
					`Variable "$%s" of type "%s" used in position expecting type "%s".`,
					value,
					varDef.Type.String(),
					expectedType.String(),
				),
			)
		}
	}
}
