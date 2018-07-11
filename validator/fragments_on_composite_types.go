package validator

import (
	"fmt"

	"github.com/vektah/gqlparser"
)

func init() {
	addRule("FragmentsOnCompositeTypes", func(observers *Events, addError addErrFunc) {
		observers.OnInlineFragment(func(walker *Walker, parentDef *gqlparser.Definition, inlineFragment *gqlparser.InlineFragment) {
			if parentDef == nil {
				return
			}

			fragmentType := walker.Schema.Types[inlineFragment.TypeCondition.Name()]
			if fragmentType == nil || fragmentType.IsCompositeType() {
				return
			}

			message := fmt.Sprintf(`Fragment cannot condition on non composite type "%s".`, inlineFragment.TypeCondition.Name())

			addError(Message(message))
		})

		observers.OnFragment(func(walker *Walker, parentDef *gqlparser.Definition, fragment *gqlparser.FragmentDefinition) {
			if parentDef == nil {
				return
			}

			if fragment.TypeCondition.Name() == "" {
				return
			} else if parentDef != nil && parentDef.IsCompositeType() {
				return
			}

			message := fmt.Sprintf(`Fragment "%s" cannot condition on non composite type "%s".`, fragment.Name, fragment.TypeCondition.Name())

			addError(Message(message))
		})
	})
}
