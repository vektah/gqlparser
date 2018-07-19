package gqlerror

import (
	"bytes"
	"fmt"
	"strconv"
)

// Error is the standard graphql error type described in https://facebook.github.io/graphql/draft/#sec-Errors
type Error struct {
	Message    string                 `json:"message"`
	Path       []interface{}          `json:"path,omitempty"`
	Locations  []Location             `json:"locations,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
	Rule       string                 `json:"-"`
}

func (err *Error) SetFile(file string) {
	if file == "" {
		return
	}
	if err.Extensions == nil {
		err.Extensions = map[string]interface{}{}
	}

	err.Extensions["file"] = file
}

type Location struct {
	Line   int `json:"line,omitempty"`
	Column int `json:"column,omitempty"`
}

type List []*Error

func (err *Error) Error() string {
	filename, _ := err.Extensions["file"].(string)
	if filename == "" {
		filename = "input"
	}

	if len(err.Locations) > 0 {
		filename += ":" + strconv.Itoa(err.Locations[0].Line)
	}

	return filename + " " + err.Message
}

func (errs List) Error() string {
	var buf bytes.Buffer
	for _, err := range errs {
		buf.WriteString(err.Error())
		buf.WriteByte('\n')
	}
	return buf.String()
}

func Errorf(message string, args ...interface{}) *Error {
	return &Error{
		Message: fmt.Sprintf(message, args...),
	}
}

func ErrorLocf(file string, line int, col int, message string, args ...interface{}) *Error {
	var extensions map[string]interface{}
	if file != "" {
		extensions = map[string]interface{}{"file": file}
	}
	return &Error{
		Message:    fmt.Sprintf(message, args...),
		Extensions: extensions,
		Locations: []Location{
			{Line: line, Column: col},
		},
	}
}
