package rules

import . "github.com/vektah/gqlparser/v2/validator/core"

type Rules struct {
	rules map[string]RuleFunc
}

func NewRules(rs ...Rule) *Rules {
	r := &Rules{
		rules: make(map[string]RuleFunc),
	}

	for _, rule := range rs {
		r.AddRule(rule.Name, rule.RuleFunc)
	}

	return r
}

func NewDefaultRules() *Rules {
	rules := []Rule{
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

func (r *Rules) AddRule(name string, ruleFunc RuleFunc) {
	if r.rules == nil {
		r.rules = make(map[string]RuleFunc)
	}
	r.rules[name] = ruleFunc
}

func (r *Rules) GetInner() map[string]RuleFunc {
	if r.rules == nil {
		return make(map[string]RuleFunc)
	}
	return r.rules
}

func (r *Rules) RemoveRule(name string) {
	if r.rules != nil {
		delete(r.rules, name)
	}
}

func (r *Rules) ReplaceRule(name string, ruleFunc RuleFunc) {
	if r.rules == nil {
		r.rules = make(map[string]RuleFunc)
	}
	r.rules[name] = ruleFunc
}
