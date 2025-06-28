package rules

import (
	"sync"

	"github.com/vektah/gqlparser/v2/validator/core"
)

// Rules manages GraphQL validation rules.
type Rules struct {
	rules *sync.Map
}

// NewRules creates a Rules instance with the specified rules.
func NewRules(rs ...core.Rule) *Rules {
	r := &Rules{
		rules: &sync.Map{},
	}

	for _, rule := range rs {
		r.AddRule(rule.Name, rule.RuleFunc)
	}

	return r
}

// NewDefaultRules creates a Rules instance containing the default GraphQL validation rule set.
func NewDefaultRules() *Rules {
	rules := []core.Rule{
		FieldsOnCorrectTypeRule,
		FragmentsOnCompositeTypesRule,
		KnownArgumentNamesRule,
		KnownDirectivesRule,
		KnownFragmentNamesRule,
		KnownRootTypeRule,
		KnownTypeNamesRule,
		LoneAnonymousOperationRule,
		MaxIntrospectionDepth,
		NoFragmentCyclesRule,
		NoUndefinedVariablesRule,
		NoUnusedFragmentsRule,
		NoUnusedVariablesRule,
		OverlappingFieldsCanBeMergedRule,
		PossibleFragmentSpreadsRule,
		ProvidedRequiredArgumentsRule,
		ScalarLeafsRule,
		SingleFieldSubscriptionsRule,
		UniqueArgumentNamesRule,
		UniqueDirectivesPerLocationRule,
		UniqueFragmentNamesRule,
		UniqueInputFieldNamesRule,
		UniqueOperationNamesRule,
		UniqueVariableNamesRule,
		ValuesOfCorrectTypeRule,
		VariablesAreInputTypesRule,
		VariablesInAllowedPositionRule,
	}

	r := NewRules(rules...)

	return r
}

// AddRule adds a rule with the specified name and rule function to the rule set.
// If a rule with the same name already exists, it will not be added.
func (r *Rules) AddRule(name string, ruleFunc core.RuleFunc) {
	if r == nil {
		// nonsensical, hopefully impossible
		return
	}
	if r.rules == nil {
		// this is probably a mistake if we get here
		r.rules = &sync.Map{}
	}
	// load only if a key doesn't already exist.
	_, _ = r.rules.LoadOrStore(name, ruleFunc)
}

// GetInner returns the internal rule map.
// If the map is not initialized, it returns an empty map.
func (r *Rules) GetInner() map[string]core.RuleFunc {
	if r.rules == nil {
		return make(map[string]core.RuleFunc)
	}

	innerCopy := make(map[string]core.RuleFunc)
	r.rules.Range(func(key, value any) bool {
		sKey, sok := key.(string)
		vKey, vok := value.(core.RuleFunc)
		if sok && vok {
			innerCopy[sKey] = vKey
		}
		return true
	})

	return innerCopy
}

// RemoveRule removes a rule with the specified name from the rule set.
// If no rule with the specified name exists, it does nothing.
func (r *Rules) RemoveRule(name string) {
	if r.rules != nil {
		r.rules.Delete(name)
	}
}

// ReplaceRule replaces a rule with the specified name with a new rule function.
// If no rule with the specified name exists, it does nothing.
func (r *Rules) ReplaceRule(name string, ruleFunc core.RuleFunc) {
	if r == nil {
		// nonsensical, hopefully impossible
		return
	}
	if r.rules == nil {
		// this is probably a mistake if we get here
		r.rules = &sync.Map{}
	}
	oldRule, ok := r.rules.Load(name)
	if ok {
		r.rules.CompareAndSwap(name, oldRule, ruleFunc)
	}
}
