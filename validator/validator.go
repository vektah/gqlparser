package validator

import (
	. "github.com/dgraph-io/gqlparser/v2/ast"
	"github.com/dgraph-io/gqlparser/v2/gqlerror"
	"sort"
)

type AddErrFunc func(options ...ErrorOption)

type ruleFunc func(observers *Events, addError AddErrFunc)

type rule struct {
	name string
	// rules will be called in the ascending order
	order int
	rule  ruleFunc
}

var rules []rule

// addRule to rule set.
// f is called once each time `Validate` is executed.
func AddRule(name string, f ruleFunc) {
	rules = append(rules, rule{name: name, rule: f})
}

// AddRuleWithOrder to rule set with an order.
// f is called once each time `Validate` is executed.
func AddRuleWithOrder(name string, order int, f ruleFunc) {
	rules = append(rules, rule{name: name, order: order, rule: f})
}

func Validate(schema *Schema, doc *QueryDocument, variables map[string]interface{}) gqlerror.List {
	var errs gqlerror.List

	observers := &Events{}
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].order < rules[j].order
	})
	for i := range rules {
		rule := rules[i]
		rule.rule(observers, func(options ...ErrorOption) {
			err := &gqlerror.Error{
				Rule: rule.name,
			}
			for _, o := range options {
				o(err)
			}
			errs = append(errs, err)
		})
	}

	Walk(schema, doc, observers, variables)
	return errs
}
