package validator

import (
	"fmt"

	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/errors"
)

func init() {
	inlineFragmentVisitors = append(inlineFragmentVisitors, inlineFragmentOnCompositeTypes)
	fragmentVisitors = append(fragmentVisitors, fragmentOnCompositeTypes)
}

func inlineFragmentOnCompositeTypes(ctx *vctx, parentDef *gqlparser.Definition, inlineFragment *gqlparser.InlineFragment) {
	if parentDef == nil {
		return
	}

	fragmentType := ctx.schema.Types[inlineFragment.TypeCondition.Name()]
	if fragmentType == nil || fragmentType.IsCompositeType() {
		return
	}

	message := fmt.Sprintf(`Fragment cannot condition on non composite type "%s".`, inlineFragment.TypeCondition.Name())

	ctx.errors = append(ctx.errors, errors.Validation{
		Message: message,
		Rule:    "FragmentsOnCompositeTypes",
	})
}

func fragmentOnCompositeTypes(ctx *vctx, parentDef *gqlparser.Definition, fragment *gqlparser.FragmentDefinition) {
	if parentDef == nil {
		return
	}

	if fragment.TypeCondition.Name() == "" {
		return
	} else if parentDef != nil && parentDef.IsCompositeType() {
		return
	}

	message := fmt.Sprintf(`Fragment "%s" cannot condition on non composite type "%s".`, fragment.Name, fragment.TypeCondition.Name())

	ctx.errors = append(ctx.errors, errors.Validation{
		Message: message,
		Rule:    "FragmentsOnCompositeTypes",
	})
}
