package validator

import "github.com/vektah/gqlparser"

func init() {
	fieldVisitors = append(fieldVisitors, scalarLeafs)
}

func scalarLeafs(ctx *vctx, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field) {
	if fieldDef == nil {
		return
	}

	fieldType := ctx.schema.Types[fieldDef.Type.Name()]
	if fieldType == nil {
		return
	}

	if fieldType.IsLeafType() && len(field.SelectionSet) > 0 {
		ctx.errors = append(ctx.errors, Error(
			Rule("ScalarLeafs"),
			Message(`Field "%s" must not have a selection since type "%s" has no subfields.`, field.Name, fieldType.Name),
		))
	}

	if !fieldType.IsLeafType() && len(field.SelectionSet) == 0 {
		ctx.errors = append(ctx.errors, Error(
			Rule("ScalarLeafs"),
			Message(`Field "%s" of type "%s" must have a selection of subfields.`, field.Name, fieldDef.Type.String()),
			Suggestf(`"%s { ... }"`, field.Name),
		))
	}
}
