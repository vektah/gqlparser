package validator

import (
	"github.com/vektah/gqlparser"
)

func init() {
	addRule("PossibleFragmentSpreads", func(observers *Events, addError addErrFunc) {

		validate := func(walker *Walker, parentDef *gqlparser.Definition, fragmentName string, emitError func()) {
			if parentDef == nil {
				return
			}

			var parentDefs []*gqlparser.Definition
			switch parentDef.Kind {
			case gqlparser.Object:
				parentDefs = []*gqlparser.Definition{parentDef}
			case gqlparser.Interface, gqlparser.Union:
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

		observers.OnInlineFragment(func(walker *Walker, parentDef *gqlparser.Definition, inlineFragment *gqlparser.InlineFragment) {
			validate(walker, parentDef, inlineFragment.TypeCondition.Name(), func() {
				addError(Message(`Fragment cannot be spread here as objects of type "%s" can never be of type "%s".`, parentDef.Name, inlineFragment.TypeCondition.Name()))
			})
		})

		observers.OnFragmentSpread(func(walker *Walker, parentDef *gqlparser.Definition, fragmentDef *gqlparser.FragmentDefinition, fragmentSpread *gqlparser.FragmentSpread) {
			if fragmentDef == nil {
				return
			}
			validate(walker, parentDef, fragmentDef.TypeCondition.Name(), func() {
				addError(Message(`Fragment "%s" cannot be spread here as objects of type "%s" can never be of type "%s".`, fragmentSpread.Name, parentDef.Name, fragmentDef.TypeCondition.Name()))
			})
		})
	})
}
