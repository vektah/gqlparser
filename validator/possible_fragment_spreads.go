package validator

import (
	"github.com/vektah/gqlparser/ast"
)

func init() {
	addRule("PossibleFragmentSpreads", func(observers *Events, addError addErrFunc) {

		validate := func(walker *Walker, parentDef *ast.Definition, fragmentName string, emitError func()) {
			if parentDef == nil {
				return
			}

			var parentDefs []*ast.Definition
			switch parentDef.Kind {
			case ast.Object:
				parentDefs = []*ast.Definition{parentDef}
			case ast.Interface, ast.Union:
				parentDefs = walker.Schema.GetPossibleTypes(parentDef)
			default:
				panic("unexpected type")
			}

			fragmentDefType := walker.Schema.Types[fragmentName]
			if fragmentDefType == nil {
				return
			}
			if !fragmentDefType.IsCompositeType() {
				// checked by FragmentsOnCompositeTypes
				return
			}
			fragmentDefs := walker.Schema.GetPossibleTypes(fragmentDefType)

			for _, fragmentDef := range fragmentDefs {
				for _, parentDef := range parentDefs {
					if parentDef.Name == fragmentDef.Name {
						return
					}
				}
			}

			emitError()
		}

		observers.OnInlineFragment(func(walker *Walker, parentDef *ast.Definition, inlineFragment *ast.InlineFragment) {
			validate(walker, parentDef, inlineFragment.TypeCondition.Name(), func() {
				addError(Message(`Fragment cannot be spread here as objects of type "%s" can never be of type "%s".`, parentDef.Name, inlineFragment.TypeCondition.Name()))
			})
		})

		observers.OnFragmentSpread(func(walker *Walker, parentDef *ast.Definition, fragmentDef *ast.FragmentDefinition, fragmentSpread *ast.FragmentSpread) {
			if fragmentDef == nil {
				return
			}
			validate(walker, parentDef, fragmentDef.TypeCondition.Name(), func() {
				addError(Message(`Fragment "%s" cannot be spread here as objects of type "%s" can never be of type "%s".`, fragmentSpread.Name, parentDef.Name, fragmentDef.TypeCondition.Name()))
			})
		})
	})
}
