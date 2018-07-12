package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	addRule("KnownTypeNames", func(observers *Events, addError addErrFunc) {
		observers.OnOperation(func(walker *Walker, operation *gqlparser.OperationDefinition) {
			for _, vdef := range operation.VariableDefinitions {
				typeName := vdef.Type.Name()
				def := walker.Schema.Types[typeName]
				if def != nil {
					continue
				}

				addError(
					Message(`Unknown type "%s".`, typeName),
				)
			}
		})

		observers.OnInlineFragment(func(walker *Walker, parentDef *gqlparser.Definition, inlineFragment *gqlparser.InlineFragment) {
			typedName := inlineFragment.TypeCondition.Name()
			if typedName == "" {
				return
			}

			def := walker.Schema.Types[typedName]
			if def != nil {
				return
			}

			addError(
				Message(`Unknown type "%s".`, typedName),
			)
		})

		observers.OnFragment(func(walker *Walker, parentDef *gqlparser.Definition, fragment *gqlparser.FragmentDefinition) {
			typeName := fragment.TypeCondition.Name()
			def := walker.Schema.Types[typeName]
			if def != nil {
				return
			}

			var possibleTypes []string
			for _, t := range walker.Schema.Types {
				possibleTypes = append(possibleTypes, t.Name)
			}

			list := SuggestListQuoted("Did you mean", typeName, possibleTypes)

			addError(
				Message(`Unknown type "%s".`, typeName),
				list,
			)
		})
	})
}
