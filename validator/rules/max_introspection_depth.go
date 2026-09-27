package rules

import (
	"github.com/vektah/gqlparser/v2/ast"
	//nolint:staticcheck // Validator rules each use dot imports for convenience.
	. "github.com/vektah/gqlparser/v2/validator/core"
)

const maxListsDepth = 3

var MaxIntrospectionDepth = Rule{
	Name: "MaxIntrospectionDepth",
	RuleFunc: func(observers *Events, addError AddErrFunc) {
		// Counts the depth of list fields in "__Type" recursively and
		// returns `true` if the limit has been reached.
		observers.OnField(func(walker *Walker, field *ast.Field) {
			if field.Name == "__schema" || field.Name == "__type" {
				c := &depthChecker{
					visitedFragments: make(map[string]bool),
					memo:             make(map[fragmentAtDepth]bool),
				}
				if c.checkDepthField(field, 0) {
					addError(
						Message(`Maximum introspection depth exceeded`),
						At(field.Position),
					)
				}
				return
			}
		})
	},
}

type fragmentAtDepth struct {
	name  string
	depth int
}

type depthChecker struct {
	visitedFragments map[string]bool
	// memo keeps the check linear in the size of the document. A fragment's
	// result depends only on the depth it is entered at, and without the
	// memo a fragment spread twice in each of n nested fragments is walked
	// 2^n times.
	memo map[fragmentAtDepth]bool
}

func (c *depthChecker) checkDepthSelectionSet(selectionSet ast.SelectionSet, depth int) bool {
	for _, child := range selectionSet {
		if field, ok := child.(*ast.Field); ok {
			if c.checkDepthField(field, depth) {
				return true
			}
		}
		if fragmentSpread, ok := child.(*ast.FragmentSpread); ok {
			if c.checkDepthFragmentSpread(fragmentSpread, depth) {
				return true
			}
		}
		if inlineFragment, ok := child.(*ast.InlineFragment); ok {
			if c.checkDepthSelectionSet(inlineFragment.SelectionSet, depth) {
				return true
			}
		}
	}
	return false
}

func (c *depthChecker) checkDepthField(field *ast.Field, depth int) bool {
	if field.Name == "fields" ||
		field.Name == "interfaces" ||
		field.Name == "possibleTypes" ||
		field.Name == "inputFields" {
		depth++
		if depth >= maxListsDepth {
			return true
		}
	}
	return c.checkDepthSelectionSet(field.SelectionSet, depth)
}

func (c *depthChecker) checkDepthFragmentSpread(
	fragmentSpread *ast.FragmentSpread,
	depth int,
) bool {
	fragmentName := fragmentSpread.Name
	if c.visitedFragments[fragmentName] {
		// Fragment cycles are handled by `NoFragmentCyclesRule`.
		return false
	}
	fragment := fragmentSpread.Definition
	if fragment == nil {
		// Missing fragments checks are handled by `KnownFragmentNamesRule`.
		return false
	}

	// Rather than following an immutable programming pattern which has
	// significant memory and garbage collection overhead, we've opted to
	// take a mutable approach for efficiency's sake. Importantly visiting a
	// fragment twice is fine, so long as you don't do one visit inside the
	// other.
	key := fragmentAtDepth{fragmentName, depth}
	if exceeded, ok := c.memo[key]; ok {
		return exceeded
	}
	c.visitedFragments[fragmentName] = true
	exceeded := c.checkDepthSelectionSet(fragment.SelectionSet, depth)
	delete(c.visitedFragments, fragmentName)
	c.memo[key] = exceeded
	return exceeded
}
