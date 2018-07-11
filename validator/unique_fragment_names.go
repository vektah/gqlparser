package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	fragmentVisitors = append(fragmentVisitors, uniqueFragmentNames)
}

// A GraphQL document is only valid if all defined fragments have unique names.
func uniqueFragmentNames(ctx *vctx, parentDef *gqlparser.Definition, fragment *gqlparser.FragmentDefinition) {
	if ctx.seenFragments[fragment.Name] {
		ctx.errors = append(ctx.errors, Error(
			Rule("UniqueFragmentNames"),
			Message(`There can be only one fragment named "%s".`, fragment.Name),
		))
	}
	ctx.seenFragments[fragment.Name] = true
}
