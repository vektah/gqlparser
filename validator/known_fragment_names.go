package validator

import (
	"fmt"

	"github.com/vektah/gqlparser"
)

func init() {
	addRule("KnownFragmentNames", func(observers *Events, addError addErrFunc) {
		observers.OnFragmentSpread(func(walker *Walker, parentDef *gqlparser.Definition, fragmentDef *gqlparser.FragmentDefinition, fragmentSpread *gqlparser.FragmentSpread) {
			if fragmentDef != nil {
				return
			}

			message := fmt.Sprintf(`Unknown fragment "%s".`, fragmentSpread.Name)
			addError(Message(message))
		})
	})
}
