package core

import (
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type AddErrFunc func(options ...ErrorOption)

type RuleFunc func(observers *Events, addError AddErrFunc)

type Rule struct {
	Name     string
	RuleFunc RuleFunc
}

type ErrorOption func(err *gqlerror.Error)