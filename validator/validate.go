package validator

import (
	. "github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/errors"
)

func Validate(schema *Schema, doc *QueryDocument) []errors.Validation {
	ctx := vctx{
		schema:        schema,
		document:      doc,
		seenFragments: map[string]bool{},
	}

	ctx.walk()

	return ctx.errors
}
