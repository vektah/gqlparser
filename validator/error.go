package validator

import (
	"fmt"

	"github.com/vektah/gqlparser/errors"
)

func Error(options ...Option) errors.Validation {
	var err errors.Validation

	for _, o := range options {
		o(&err)
	}

	return err
}

type Option func(err *errors.Validation)

func Rule(rule string) Option {
	return func(err *errors.Validation) {
		err.Rule = rule
	}
}

func Message(msg string, args ...interface{}) Option {
	return func(err *errors.Validation) {
		err.Message = fmt.Sprintf(msg, args...)
	}
}

func SuggestList(typed string, suggestions []string) Option {
	suggested := suggestionList(typed, suggestions)
	return func(err *errors.Validation) {
		if len(suggested) > 0 {
			err.Message += " Did you mean " + quotedOrList(suggested...) + "?"
		}
	}
}

func Suggestf(suggestion string, args ...interface{}) Option {
	return func(err *errors.Validation) {
		err.Message += " Did you mean " + fmt.Sprintf(suggestion, args...) + "?"
	}
}
