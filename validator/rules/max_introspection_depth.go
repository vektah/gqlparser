package rules

import (
	"github.com/vektah/gqlparser/v2/ast"
	//nolint:staticcheck // Validator rules each use dot imports for convenience.
	. "github.com/vektah/gqlparser/v2/validator/core"
)

const maxListsDepth = 3

// MaxIntrospectionDepth reports an introspection selection that nests __Type's list
// fields past a fixed depth.
//
// The rule assumes NoFragmentCyclesRule runs alongside it, as it does in the default rule
// set. Selected on its own, it can pass a document whose fragments form a cycle where an
// unmemoized walk would have reported a violation; depthChecker explains why.
var MaxIntrospectionDepth = Rule{
	Name: "MaxIntrospectionDepth",
	RuleFunc: func(observers *Events, addError AddErrFunc) {
		// One checker per document: a fragment's result at a given depth is the same
		// whichever introspection root reached it, so a document with many roots that
		// spread the same fragments checks each of them once rather than once per root.
		c := &depthChecker{
			visitedFragments: make(map[string]bool),
			memo:             make(map[fragmentAtDepth]bool),
		}

		// Counts the depth of list fields in "__Type" recursively and
		// returns `true` if the limit has been reached.
		observers.OnField(func(walker *Walker, field *ast.Field) {
			if field.Name == "__schema" || field.Name == "__type" {
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

// depthChecker walks a document's introspection fields, reusing each fragment's result
// rather than following every path to it.
//
// A fragment's result depends on the depth it is entered at and on which fragments are
// already being visited, because visitedFragments cuts a fragment that is reached again
// while it is still on the stack. Only a cycle can reach a fragment that is still being
// visited, so for an acyclic document the cut never happens and the entry depth alone
// identifies the result.
//
// On a cyclic document the memo can therefore answer with a result reached under a
// different set of visited fragments, and miss a depth error the uncached walk reports.
// That is accepted: under the default rule set NoFragmentCyclesRule rejects every such
// document, so nothing that used to fail validation now passes. A rule set that selects
// MaxIntrospectionDepth without NoFragmentCyclesRule does lose the error, which is why the
// rule documents the dependency. Skipping the memo for a fragment whose walk cut a cycle
// would keep the result exact, but it would also restore the exponential walk for cyclic
// documents, and the cost of that walk is the denial of service this memo exists to
// prevent.
//
// graphql-js does not memoize here and pays the exponential walk, so this is a deliberate
// divergence from the reference implementation.
type depthChecker struct {
	visitedFragments map[string]bool
	// memo keeps the walk linear: without it, a fragment spread twice in each of n nested
	// fragments is walked 2^n times.
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

	key := fragmentAtDepth{fragmentName, depth}
	if exceeded, ok := c.memo[key]; ok {
		return exceeded
	}

	// Rather than following an immutable programming pattern which has
	// significant memory and garbage collection overhead, we've opted to
	// take a mutable approach for efficiency's sake. Importantly visiting a
	// fragment twice is fine, so long as you don't do one visit inside the
	// other.
	c.visitedFragments[fragmentName] = true
	defer delete(c.visitedFragments, fragmentName)

	exceeded := c.checkDepthSelectionSet(fragment.SelectionSet, depth)
	c.memo[key] = exceeded
	return exceeded
}
