package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	addRule("UniqueInputFieldNames", func(observers *Events, addError addErrFunc) {
		observers.OnValue(func(walker *Walker, fieldType gqlparser.Type, def *gqlparser.Definition, value gqlparser.Value) {
			object, isObject := value.(gqlparser.ObjectValue)
			if !isObject {
				return
			}

			seen := map[string]bool{}
			for _, field := range object {
				if seen[field.Name] {
					addError(
						Message(`There can be only one input field named "%s".`, field.Name),
					)
				}
				seen[field.Name] = true
			}
		})
	})
}
