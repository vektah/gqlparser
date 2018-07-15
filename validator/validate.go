package validator

import (
	. "github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/errors"
)

type AddErrFunc func(options ...ErrorOption)

type ruleFunc func(observers *Events, addError AddErrFunc)

type rule struct {
	name string
	rule ruleFunc
}

var rules []rule

// addRule to rule set.
// f is called once each time `Validate` is executed.
func AddRule(name string, f ruleFunc) {
	rules = append(rules, rule{name: name, rule: f})
}

func Validate(schema *Schema, doc *QueryDocument) []errors.Validation {
	var errs []errors.Validation

	observers := &Events{}
	for i := range rules {
		rule := rules[i]
		rule.rule(observers, func(options ...ErrorOption) {
			err := errors.Validation{
				Rule: rule.name,
			}
			for _, o := range options {
				o(&err)
			}
			errs = append(errs, err)
		})
	}

	Walk(schema, doc, observers)
	return errs
}
