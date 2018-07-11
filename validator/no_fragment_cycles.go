package validator

import (
	"fmt"
	"strings"

	"github.com/vektah/gqlparser"
)

func init() {
	addRule("NoFragmentCycles", func(observers *Events, addError addErrFunc) {
		visitedFrags := make(map[string]bool)

		observers.OnFragment(func(walker *Walker, parentDef *gqlparser.Definition, fragment *gqlparser.FragmentDefinition) {
			var spreadPath []gqlparser.FragmentSpread
			spreadPathIndexByName := make(map[string]int)

			var recursive func(fragment *gqlparser.FragmentDefinition)
			recursive = func(fragment *gqlparser.FragmentDefinition) {
				if visitedFrags[fragment.Name] {
					return
				}

				visitedFrags[fragment.Name] = true

				spreadNodes := getFragmentSpreads(fragment.SelectionSet)
				if len(spreadNodes) == 0 {
					return
				}
				spreadPathIndexByName[fragment.Name] = len(spreadPath)

				for _, spreadNode := range spreadNodes {
					spreadName := spreadNode.Name

					cycleIndex, ok := spreadPathIndexByName[spreadName]

					spreadPath = append(spreadPath, spreadNode)
					if !ok {
						spreadFragment := walker.Document.GetFragment(spreadName)
						if spreadFragment != nil {
							recursive(spreadFragment)
						}
					} else {
						cyclePath := spreadPath[cycleIndex : len(spreadPath)-1]
						var fragmentNames []string
						for _, fs := range cyclePath {
							fragmentNames = append(fragmentNames, fs.Name)
						}
						var via string
						if len(fragmentNames) != 0 {
							via = fmt.Sprintf(" via %s", strings.Join(fragmentNames, ", "))
						}
						addError(Message(`Cannot spread fragment "%s" within itself%s.`, spreadName, via))
					}

					spreadPath = spreadPath[:len(spreadPath)-1]
				}

				delete(spreadPathIndexByName, fragment.Name)
			}

			recursive(fragment)
		})
	})
}

func getFragmentSpreads(node gqlparser.SelectionSet) []gqlparser.FragmentSpread {
	var spreads []gqlparser.FragmentSpread

	setsToVisit := []gqlparser.SelectionSet{node}

	for len(setsToVisit) != 0 {
		set := setsToVisit[len(setsToVisit)-1]
		setsToVisit = setsToVisit[:len(setsToVisit)-1]

		for _, selection := range set {
			switch selection := selection.(type) {
			case gqlparser.FragmentSpread:
				spreads = append(spreads, selection)
			case gqlparser.Field:
				setsToVisit = append(setsToVisit, selection.SelectionSet)
			case gqlparser.InlineFragment:
				setsToVisit = append(setsToVisit, selection.SelectionSet)
			}
		}
	}

	return spreads
}
