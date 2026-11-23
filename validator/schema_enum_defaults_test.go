package validator_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestLoadSchemaEnumDefaultLiterals(t *testing.T) {
	for _, tc := range []struct {
		name, sdl string
		reject    bool
	}{
		{"input field", `input Bar { value: Foo = "BAR" }`, true},
		{"list", `input Bar { values: [Foo!] = ["BAR"] }`, true},
		{"singleton list", `input Bar { values: [Foo] = "BAR" }`, true},
		{"nested list", `input Bar { values: [[Foo]] = [["BAR"]] }`, true},
		{"block string", `input Bar { value: Foo = """BAR""" }`, true},
		{"field argument", `type Query { field(arg: Foo = "BAR"): Int }`, true},
		{"interface field argument", `interface Baz { field(arg: Foo = "BAR"): Int }`, true},
		{"directive argument", `directive @bar(arg: Foo = "BAR") on FIELD_DEFINITION`, true},
		{"object default", `input Bar { value: Foo } type Query { field(arg: Bar = {value: "BAR"}): Int }`, true},
		{"enum literals", `input Bar { value: Foo = BAR values: [Foo!] = [BAR] }`, false},
		{"nullable enum defaults", `input Bar { value: Foo = null values: [Foo] = [null] }`, false},
		{"singleton enum list", `input Bar { values: [Foo] = BAR }`, false},
		{"string default", `input Bar { value: String = "BAR" }`, false},
		{"custom scalar default", `scalar Baz input Bar { value: Baz = "BAR" }`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := gqlparser.LoadSchema(
				&ast.Source{Name: "schema.graphql", Input: "enum Foo { BAR }\n" + tc.sdl},
			)
			if tc.reject {
				require.ErrorContains(t, err, `Enum "Foo" cannot represent non-enum value`)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
