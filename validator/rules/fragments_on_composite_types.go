package validator

import (
	"fmt"

	"github.com/vektah/gqlparser/ast"
	. "github.com/vektah/gqlparser/validator"
)

func init() {
	AddRule("FragmentsOnCompositeTypes", func(observers *Events, addError AddErrFunc) {
		observers.OnInlineFragment(func(walker *Walker, inlineFragment *ast.InlineFragment) {
			fragmentType := walker.Schema.Types[inlineFragment.TypeCondition.Name()]
			if fragmentType == nil || fragmentType.IsCompositeType() {
				return
			}

			message := fmt.Sprintf(`Fragment cannot condition on non composite type "%s".`, inlineFragment.TypeCondition.Name())

			addError(Message(message))
		})

		observers.OnFragment(func(walker *Walker, fragment *ast.FragmentDefinition) {
			if fragment.Definition == nil || fragment.TypeCondition.Name() == "" || fragment.Definition.IsCompositeType() {
				return
			}

			message := fmt.Sprintf(`Fragment "%s" cannot condition on non composite type "%s".`, fragment.Name, fragment.TypeCondition.Name())

			addError(Message(message))
		})
	})
}
