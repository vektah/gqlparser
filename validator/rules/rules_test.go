package rules_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/validator/core"
	"github.com/vektah/gqlparser/v2/validator/rules"
)

// newDummyRule returns a dummy Rule for testing purposes.
// The RuleFunc is a no-op; only the name is relevant for these tests.
func newDummyRule(name string) core.Rule {
	return core.Rule{
		Name:     name,
		RuleFunc: func(*core.Events, core.AddErrFunc) {},
	}
}

// TestNewRules ensures that NewRules registers the provided Rules in the internal
// map and keeps ruleNameKeys in the expected order.
func TestNewRules(t *testing.T) {
	r1 := newDummyRule("FirstRule")
	r2 := newDummyRule("SecondRule")

	rs := rules.NewRules(r1, r2)
	inner := rs.GetInner()

	require.Len(t, inner, 2)
	require.NotNil(t, inner["FirstRule"])
	require.NotNil(t, inner["SecondRule"])
	require.Equal(t, []string{"FirstRule", "SecondRule"}, rs.GetRuleNameKeys())
}

// TestAddRuleDuplicate confirms that calling AddRule twice with the same name
// does not create duplicate entries.
func TestAddRuleDuplicate(t *testing.T) {
	rs := &rules.Rules{}

	rs.AddRule("DupRule", func(*core.Events, core.AddErrFunc) {})
	rs.AddRule("DupRule", func(*core.Events, core.AddErrFunc) {})

	inner := rs.GetInner()

	require.Len(t, inner, 1)
	require.Equal(t, []string{"DupRule"}, rs.GetRuleNameKeys())
}

// TestRemoveRule verifies that RemoveRule deletes the entry from both the internal
// map and ruleNameKeys slice.
func TestRemoveRule(t *testing.T) {
	rs := &rules.Rules{}

	rs.AddRule("RemoveMe", func(*core.Events, core.AddErrFunc) {})
	rs.RemoveRule("RemoveMe")

	inner := rs.GetInner()

	require.Empty(t, inner)
	require.NotContains(t, rs.GetRuleNameKeys(), "RemoveMe")
}

// TestReplaceRule checks that ReplaceRule actually swaps out the RuleFunc for an
// existing rule.
func TestReplaceRule(t *testing.T) {
	rs := &rules.Rules{}

	oldFunc := func(*core.Events, core.AddErrFunc) {}
	rs.AddRule("Target", oldFunc)

	newFunc := func(*core.Events, core.AddErrFunc) {}
	rs.ReplaceRule("Target", newFunc)

	inner := rs.GetInner()

	require.Len(t, inner, 1)
	require.Equal(t, reflect.ValueOf(newFunc).Pointer(), reflect.ValueOf(inner["Target"]).Pointer())
}

// TestGetInnerNilSafety guarantees that GetInner is safe on a nil receiver and on
// an empty Rules struct, returning sensible defaults without panicking.
func TestGetInnerNilSafety(t *testing.T) {
	var rs *rules.Rules

	require.Nil(t, rs)
	require.Nil(t, rs.GetInner())

	rs2 := &rules.Rules{}
	inner := rs2.GetInner()
	require.Empty(t, inner)
}

// TestNewDefaultRules asserts that NewDefaultRules returns a non-empty rule set.
func TestNewDefaultRules(t *testing.T) {
	rs := rules.NewDefaultRules()
	inner := rs.GetInner()

	require.NotEmpty(t, inner)
}

// TestGetInnerReturnsCopy confirms that GetInner returns a copy of the internal map.
func TestGetInnerReturnsCopy(t *testing.T) {
	rs := rules.NewRules(core.Rule{Name: "DummyRule", RuleFunc: func(*core.Events, core.AddErrFunc) {}})

	first := rs.GetInner()
	delete(first, "DummyRule")

	second := rs.GetInner()
	if _, ok := second["DummyRule"]; !ok {
		t.Fatalf("modifying the returned map affected internal state")
	}
}
