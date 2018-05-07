package parser

import (
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/andreyvit/diff"
	yaml "gopkg.in/yaml.v2"
)

func TestQueryDocument(t *testing.T) {
	b, err := ioutil.ReadFile("queryspec.yml")
	if err != nil {
		panic(err)
	}
	var tests testcase
	err = yaml.Unmarshal(b, &tests)
	if err != nil {
		t.Errorf("unable to load spec.yml: %s", err.Error())
		return
	}

	for name, testcase := range tests {
		t.Run(name, func(t *testing.T) {
			for name, test := range testcase {
				t.Run(name, test.Run)
			}
		})
	}
}

type testcase map[string]map[string]test

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
	AST string
}

func (test *test) Run(t *testing.T) {
	for name, child := range test.Subtests {
		t.Run(name, child.Run)
	}

	doc, err := ParseQuery(test.Input)

	t.Logf("input: %s", strconv.Quote(test.Input))
	if err != nil {
		t.Logf("error: %s", err.Error())
	}
	ast := Dump(doc)
	t.Logf("ast:\n%s", ast)

	if test.Error.Message == "" {
		if err != nil {
			t.Errorf("unexpected error %s", err.Error())
		}
	} else if err == nil {
		t.Errorf("missing error\nexpected: %s\ngot:      <nil>", test.Error.Message)
	} else {
		if err.Message != test.Error.Message {
			t.Errorf("wrong error returned\nexpected: %s\ngot:      %s", test.Error.Message, err.Message)
		}

		if err.Locations[0].Column != test.Error.Location.Column || err.Locations[0].Line != test.Error.Location.Line {
			t.Errorf(
				"wrong error location:\nexpected: line %d column %d\ngot:      line %d column %d",
				test.Error.Location.Line,
				test.Error.Location.Column,
				err.Locations[0].Line,
				err.Locations[0].Column,
			)
		}
	}

	if test.AST != "" && ast != test.AST {
		diff := diff.LineDiff(test.AST, ast)
		if diff != "" {
			t.Errorf("AST mismatch:\n%s", diff)
		}
	}

}
