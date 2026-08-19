package rules

import (
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
)

func Test_sameArguments(t *testing.T) {
	tests := map[string]struct {
		args   func() (args1, args2 []*ast.Argument)
		result bool
	}{
		"both argument lists empty": {
			args: func() (args1 []*ast.Argument, args2 []*ast.Argument) {
				return nil, nil
			},
			result: true,
		},
		"args 1 empty, args 2 not": {
			args: func() (args1 []*ast.Argument, args2 []*ast.Argument) {
				return nil, []*ast.Argument{
					{
						Name:     "thing",
						Value:    &ast.Value{Raw: "a thing"},
						Position: &ast.Position{},
					},
				}
			},
			result: false,
		},
		"args 2 empty, args 1 not": {
			args: func() (args1 []*ast.Argument, args2 []*ast.Argument) {
				return []*ast.Argument{
					{
						Name:     "thing",
						Value:    &ast.Value{Raw: "a thing"},
						Position: &ast.Position{},
					},
				}, nil
			},
			result: false,
		},
		"args 1 mismatches args 2 names": {
			args: func() (args1 []*ast.Argument, args2 []*ast.Argument) {
				return []*ast.Argument{
						{
							Name:     "thing1",
							Value:    &ast.Value{Raw: "1 thing"},
							Position: &ast.Position{},
						},
					},
					[]*ast.Argument{
						{
							Name:     "thing2",
							Value:    &ast.Value{Raw: "2 thing"},
							Position: &ast.Position{},
						},
					}
			},
			result: false,
		},
		"args 1 mismatches args 2 values": {
			args: func() (args1 []*ast.Argument, args2 []*ast.Argument) {
				return []*ast.Argument{
						{
							Name:     "thing1",
							Value:    &ast.Value{Raw: "1 thing"},
							Position: &ast.Position{},
						},
					},
					[]*ast.Argument{
						{
							Name:     "thing1",
							Value:    &ast.Value{Raw: "2 thing"},
							Position: &ast.Position{},
						},
					}
			},
			result: false,
		},
		"args 1 matches args 2 names and values": {
			args: func() (args1 []*ast.Argument, args2 []*ast.Argument) {
				return []*ast.Argument{
						{
							Name:     "thing1",
							Value:    &ast.Value{Raw: "1 thing"},
							Position: &ast.Position{},
						},
					},
					[]*ast.Argument{
						{
							Name:     "thing1",
							Value:    &ast.Value{Raw: "1 thing"},
							Position: &ast.Position{},
						},
					}
			},
			result: true,
		},
		"args 1 matches args 2 names and values where multiple exist in various orders": {
			args: func() (args1 []*ast.Argument, args2 []*ast.Argument) {
				return []*ast.Argument{
						{
							Name:     "thing1",
							Value:    &ast.Value{Raw: "1 thing"},
							Position: &ast.Position{},
						},
						{
							Name:     "thing2",
							Value:    &ast.Value{Raw: "2 thing"},
							Position: &ast.Position{},
						},
					},
					[]*ast.Argument{
						{
							Name:     "thing1",
							Value:    &ast.Value{Raw: "1 thing"},
							Position: &ast.Position{},
						},
						{
							Name:     "thing2",
							Value:    &ast.Value{Raw: "2 thing"},
							Position: &ast.Position{},
						},
					}
			},
			result: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			args1, args2 := tc.args()

			resp := sameArguments(args1, args2)

			if resp != tc.result {
				t.Fatalf("Expected %t got %t", tc.result, resp)
			}
		})
	}
}

func Test_sameValue(t *testing.T) {
	str := func(raw string) *ast.Value {
		return &ast.Value{Kind: ast.StringValue, Raw: raw}
	}
	object := func(fields ...*ast.ChildValue) *ast.Value {
		return &ast.Value{Kind: ast.ObjectValue, Children: ast.ChildValueList(fields)}
	}
	field := func(name string, v *ast.Value) *ast.ChildValue {
		return &ast.ChildValue{Name: name, Value: v}
	}
	list := func(elems ...*ast.Value) *ast.Value {
		children := make(ast.ChildValueList, len(elems))
		for i, e := range elems {
			children[i] = &ast.ChildValue{Value: e}
		}
		return &ast.Value{Kind: ast.ListValue, Children: children}
	}

	tests := map[string]struct {
		value1, value2 *ast.Value
		result         bool
	}{
		"identical scalars": {str("DE"), str("DE"), true},
		"different scalars":  {str("DE"), str("US"), false},
		// Regression for #108: object arguments are stored in Children with an empty
		// Raw, so differing objects used to compare equal and merge incorrectly.
		"objects with differing field values": {
			object(field("code", str("DE"))),
			object(field("code", str("US"))),
			false,
		},
		"identical objects": {
			object(field("code", str("DE"))),
			object(field("code", str("DE"))),
			true,
		},
		// Input object field order is not significant.
		"objects with same fields in different order": {
			object(field("code", str("DE")), field("name", str("Germany"))),
			object(field("name", str("Germany")), field("code", str("DE"))),
			true,
		},
		"objects with different field count": {
			object(field("code", str("DE"))),
			object(field("code", str("DE")), field("name", str("Germany"))),
			false,
		},
		// Regression for #108: same for list arguments.
		"lists with differing elements": {
			list(str("DE")),
			list(str("US")),
			false,
		},
		"identical lists": {
			list(str("DE"), str("US")),
			list(str("DE"), str("US")),
			true,
		},
		// List element order is significant.
		"lists with same elements in different order": {
			list(str("DE"), str("US")),
			list(str("US"), str("DE")),
			false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := sameValue(tc.value1, tc.value2); got != tc.result {
				t.Fatalf("Expected %t got %t", tc.result, got)
			}
		})
	}
}
