package validator

import (
	"fmt"
	"sort"

	"github.com/vektah/gqlparser"
)

func init() {
	addRule("FieldsOnCorrectType", func(observers *Events, addError addErrFunc) {
		observers.OnField(func(walker *Walker, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field) {
			if parentDef == nil {
				return
			}

			if fieldDef != nil {
				return
			}

			message := fmt.Sprintf(`Cannot query field "%s" on type "%s".`, field.Name, parentDef.Name)

			if suggestedTypeNames := getSuggestedTypeNames(walker, parentDef, field.Name); suggestedTypeNames != nil {
				message += " Did you mean to use an inline fragment on " + quotedOrList(suggestedTypeNames...) + "?"
			} else if suggestedFieldNames := getSuggestedFieldNames(parentDef, field.Name); suggestedFieldNames != nil {
				message += " Did you mean " + quotedOrList(suggestedFieldNames...) + "?"
			}

			addError(Message(message))
		})
	})
}

// Go through all of the implementations of type, as well as the interfaces
// that they implement. If any of those types include the provided field,
// suggest them, sorted by how often the type is referenced,  starting
// with Interfaces.
func getSuggestedTypeNames(walker *Walker, parent *gqlparser.Definition, name string) []string {
	if !parent.IsAbstractType() {
		return nil
	}

	var suggestedObjectTypes []string
	var suggestedInterfaceTypes []string
	interfaceUsageCount := map[string]int{}

	for _, possibleType := range walker.Schema.GetPossibleTypes(parent) {
		field := possibleType.Field(name)
		if field == nil {
			continue
		}

		suggestedObjectTypes = append(suggestedObjectTypes, possibleType.Name)

		for _, possibleInterface := range possibleType.Interfaces {
			interfaceField := walker.Schema.Types[possibleInterface.Name()]
			if interfaceField != nil && interfaceField.Field(name) != nil {
				if interfaceUsageCount[possibleInterface.Name()] == 0 {
					suggestedInterfaceTypes = append(suggestedInterfaceTypes, possibleInterface.Name())
				}
				interfaceUsageCount[possibleInterface.Name()]++
			}
		}
	}

	sort.SliceStable(suggestedInterfaceTypes, func(i, j int) bool {
		return interfaceUsageCount[suggestedInterfaceTypes[i]] > interfaceUsageCount[suggestedInterfaceTypes[j]]
	})

	return append(suggestedInterfaceTypes, suggestedObjectTypes...)
}

// For the field name provided, determine if there are any similar field names
// that may be the result of a typo.
func getSuggestedFieldNames(parent *gqlparser.Definition, name string) []string {
	if parent.Kind != gqlparser.Object && parent.Kind != gqlparser.Interface {
		return nil
	}

	var possibleFieldNames []string
	for _, field := range parent.Fields {
		possibleFieldNames = append(possibleFieldNames, field.Name)
	}

	return suggestionList(name, possibleFieldNames)
}
