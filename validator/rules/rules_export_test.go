package rules

// GetRuleNameKeys is a test helper to access the private field ruleNameKeys.
// This returns a copy of the ruleNameKeys slice, not the original slice.
func (r *Rules) GetRuleNameKeys() []string {
	keys := make([]string, len(r.ruleNameKeys))
	copy(keys, r.ruleNameKeys)

	return keys
}
