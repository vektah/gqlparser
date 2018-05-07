package lexer

import (
	"io/ioutil"
	"strconv"
	"strings"
	"testing"

	"github.com/vektah/graphql-parser"
	"gopkg.in/yaml.v2"
)

func TestLexer(t *testing.T) {
	b, err := ioutil.ReadFile("spec.yml")
	if err != nil {
		panic(err)
	}
	var tests testcase
	err = yaml.Unmarshal(b, &tests)
	if err != nil {
		t.Errorf("unable to load spec.yml: %s", err.Error())
		return
	}

	for name, test := range tests {
		t.Run(name, test.Run)
	}
}

type testcase map[string]test

type test struct {
	Input    string
	Subtests map[string]test
	Error    struct {
		Message  string
		Location struct {
			Line   int
			Column int
		}
	}
	Tokens []struct {
		Kind   string
		Value  string
		Start  int
		End    int
		Line   int
		Column int
	}
}

func (test *test) Run(t *testing.T) {
	for name, child := range test.Subtests {
		t.Run(name, child.Run)
	}

	if test.Input == "" {
		return
	}

	l := New(test.Input)
	var tokens []Token
	var error *graphql_parser.Error

	for {
		tok, err := l.ReadToken()

		if err != nil {
			error = err
			break
		}

		if tok.Kind == EOF {
			break
		}

		tokens = append(tokens, tok)
	}

	t.Logf("input: %s", strconv.Quote(test.Input))
	if error != nil {
		t.Logf("error: %s", error.Error())
	}
	t.Log("tokens: ")
	for _, tok := range tokens {
		t.Logf("  - %s", tok.String())
	}
	t.Logf("  - <EOF>")

	if test.Error.Message == "" {
		if error != nil {
			t.Errorf("unexpected error %s", error.Error())
		}
	} else if error == nil {
		t.Errorf("expected error %s but got none", test.Error.Message)
	} else {
		if error.Message != test.Error.Message {
			t.Errorf("wrong error returned\nexpected: %s\ngot:      %s", test.Error.Message, error.Message)
		}

		if error.Locations[0].Column != test.Error.Location.Column || error.Locations[0].Line != test.Error.Location.Line {
			t.Errorf(
				"wrong error location:\nexpected: line %d column %d\ngot:      line %d column %d",
				test.Error.Location.Line,
				test.Error.Location.Column,
				error.Locations[0].Line,
				error.Locations[0].Column,
			)
		}
	}

	if len(test.Tokens) != len(tokens) {
		var tokensStr []string
		for _, t := range tokens {
			tokensStr = append(tokensStr, t.String())
		}
		t.Errorf("token count mismatch, got: \n%s", strings.Join(tokensStr, "\n"))
	} else {
		for i, tok := range tokens {
			expected := test.Tokens[i]

			if !strings.EqualFold(strings.Replace(expected.Kind, "_", "", -1), tok.Kind.Name()) {
				t.Errorf("token[%d].kind should be %s, was %s", i, expected.Kind, tok.Kind.Name())
			}
			if expected.Value != "undefined" && expected.Value != tok.Value {
				t.Errorf("token[%d].value incorrect\nexpected: %s\ngot:      %s", i, strconv.Quote(expected.Value), strconv.Quote(tok.Value))
			}
			if expected.Start != 0 && expected.Start != tok.Start {
				t.Errorf("token[%d].start should be %d, was %d", i, expected.Start, tok.Start)
			}
			if expected.End != 0 && expected.End != tok.End {
				t.Errorf("token[%d].end should be %d, was %d", i, expected.End, tok.End)
			}
			if expected.Line != 0 && expected.Line != tok.Line {
				t.Errorf("token[%d].line should be %d, was %d", i, expected.Line, tok.Line)
			}
			if expected.Column != 0 && expected.Column != tok.Column {
				t.Errorf("token[%d].column should be %d, was %d", i, expected.Column, tok.Column)
			}
		}
	}
}
