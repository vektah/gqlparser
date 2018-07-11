package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	addRule("ScalarLeafs", func(observers *Events, addError addErrFunc) {
		observers.OnField(func(walker *Walker, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field) {
			if fieldDef == nil {
				return
			}

			fieldType := walker.Schema.Types[fieldDef.Type.Name()]
			if fieldType == nil {
				return
			}

			if fieldType.IsLeafType() && len(field.SelectionSet) > 0 {
				addError(
					Message(`Field "%s" must not have a selection since type "%s" has no subfields.`, field.Name, fieldType.Name),
				)
			}

			if !fieldType.IsLeafType() && len(field.SelectionSet) == 0 {
				addError(
					Message(`Field "%s" of type "%s" must have a selection of subfields.`, field.Name, fieldDef.Type.String()),
					Suggestf(`"%s { ... }"`, field.Name),
				)
			}
		})
	})
}
