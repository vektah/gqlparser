package validator

import (
	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/errors"
)

func init() {
	fieldVisitors = append(fieldVisitors, fieldsOnCorrectType)
}

func fieldsOnCorrectType(ctx *vctx, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field) {
	if parentDef == nil {
		return
	}

	if fieldDef != nil {
		return
	}

	ctx.errors = append(ctx.errors, errors.Validation{Message: "Df"})
}
