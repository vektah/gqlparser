package lexer

import (
	"testing"

	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/spec"
)

func TestLexer(t *testing.T) {
	spec.Test(t, "../spec/lexer.yml", func(t *testing.T, input string) spec.Spec {
		l := New(&ast.Source{Input: input, Name: "spec"})

		ret := spec.Spec{}
		for {
			tok, err := l.ReadToken()

			if err != nil {
				ret.Error = err
				break
			}

			if tok.Kind == EOF {
				break
			}

			ret.Tokens = append(ret.Tokens, spec.Token{
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
