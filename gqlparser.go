//go:generate go run ./inliner/inliner.go

package gqlparser

import (
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/gqlerror"
	"github.com/vektah/gqlparser/parser"
	"github.com/vektah/gqlparser/validator"

	_ "github.com/vektah/gqlparser/validator/rules"
)

func LoadSchema(str ...*ast.Source) (*ast.Schema, *gqlerror.Error) {
	return parser.LoadSchema(append([]*ast.Source{Prelude}, str...)...)
}

func MustLoadSchema(str ...*ast.Source) *ast.Schema {
	s, err := parser.LoadSchema(append([]*ast.Source{Prelude}, str...)...)
	if err != nil {
		panic(err)
	}
	return s
}

func LoadQuery(schema *ast.Schema, str string) (*ast.QueryDocument, gqlerror.List) {
	query, err := parser.ParseQuery(&ast.Source{Input: str})
	if err != nil {
		return nil, gqlerror.List{err}
	}
	errs := validator.Validate(schema, query)
	if errs != nil {
		return nil, errs
	}

	return query, nil
}
