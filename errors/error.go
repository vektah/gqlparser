package errors

import (
	"bytes"
	"fmt"
)

type Syntax struct {
	Message   string     `json:"message"`
	Locations []Location `json:"locations,omitempty"`
}

type Validation struct {
	Message   string     `json:"message"`
	Locations []Location `json:"locations,omitempty"`
	Rule      string     `json:"-"`
}

type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type ValidationErrors []Validation

func (err Syntax) Error() string {
	str := fmt.Sprintf("Syntax Error: %s", err.Message)
	for _, loc := range err.Locations {
		str += fmt.Sprintf(" (line %d, column %d)", loc.Line, loc.Column)
	}
	return str
}

func (err Validation) Error() string {
	str := fmt.Sprintf("Validation Error: %s", err.Message)
	for _, loc := range err.Locations {
		str += fmt.Sprintf(" (line %d, column %d)", loc.Line, loc.Column)
	}
	return str
}

func (errs ValidationErrors) Error() string {
	var buf bytes.Buffer
	for _, err := range errs {
		buf.WriteString(err.Error())
		buf.WriteByte('\n')
	}
	return buf.String()
}
