package validator

import (
	"github.com/vektah/gqlparser/ast"
	. "github.com/vektah/gqlparser/validator"
)

func init() {
	AddRule("ScalarLeafs", func(observers *Events, addError AddErrFunc) {
		observers.OnField(func(walker *Walker, parentDef *ast.Definition, fieldDef *ast.FieldDefinition, field *ast.Field) {
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
