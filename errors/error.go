package errors

import "fmt"

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

func (err Syntax) Error() string {
	str := fmt.Sprintf("Syntax Error: %s", err.Message)
	for _, loc := range err.Locations {
		str += fmt.Sprintf(" (line %d, column %d)", loc.Line, loc.Column)
	}
	return str
}
