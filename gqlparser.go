package gqlparser

import (
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/parser"
	"github.com/vektah/gqlparser/validator"
)

func LoadSchema(str string) (*ast.Schema, error) {
	return parser.LoadSchema(str)
}

func LoadQuery(schema *ast.Schema, str string) (*ast.QueryDocument, error) {
	query, err := parser.ParseQuery(str)
	if err != nil {
		return nil, err
	}
	errs := validator.Validate(schema, query)
	if errs != nil {
		return nil, errs
	}

	return query, nil
}
