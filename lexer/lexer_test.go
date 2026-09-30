package lexer_test

import (
	"errors"
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/lexer"
	"github.com/vektah/gqlparser/v2/parser/testrunner"
)

func TestLexer(t *testing.T) {
	testrunner.Test(t, "lexer_test.yml", func(t *testing.T, input string) testrunner.Spec {
		l := lexer.New(&ast.Source{Input: input, Name: "spec"})

		ret := testrunner.Spec{}
		for {
			tok, err := l.ReadToken()
			if err != nil {
				ret.Error = func() *testrunner.SpecError {
					target := &testrunner.SpecError{}
					_ = errors.As(err, &target.Error)
					return target
				}()
				break
			}

			if tok.Kind == lexer.EOF {
				break
			}

			ret.Tokens = append(ret.Tokens, testrunner.Token{
				Kind:   tok.Kind.Name(),
				Value:  tok.Value,
				Line:   tok.Pos.Line,
				Column: tok.Pos.Column,
				Start:  tok.Pos.Start,
				End:    tok.Pos.End,
				Src:    tok.Pos.Src.Name,
			})
		}

		return ret
	})
}
