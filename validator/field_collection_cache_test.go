package validator_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
)

const fieldCollectionSchema = `
type Query { item: Item otherItem: Item result: Result }
interface Item { id: ID! name: String child: Item value(arg: Int): String }
type A implements Item { id: ID! name: String child: Item value(arg: Int): String a: String }
type B implements Item { id: ID! name: String child: Item value(arg: Int): String b: String }
union Result = A | B
`

func TestFieldCollectionCacheValidation(t *testing.T) {
	schema := gqlparser.MustLoadSchema(&ast.Source{Input: fieldCollectionSchema})
	tests := []struct {
		name    string
		query   string
		message string
	}{
		{
			name: "repeated fragment on different paths",
			query: `{
				item { ...Shared ...Shared }
				otherItem { ...Shared }
			} fragment Shared on Item { id child { id } }`,
		},
		{
			name: "nested alias conflict between fragments",
			query: `{ item { ...Left ...Right } }
				fragment Left on Item { child { same: id } }
				fragment Right on Item { child { same: name } }`,
			message: `subfields "same" conflict because "id" and "name" are different fields`,
		},
		{
			name: "argument conflict between fragments",
			query: `{ item { ...Left ...Right } }
				fragment Left on Item { value(arg: 1) }
				fragment Right on Item { value(arg: 2) }`,
			message: "they have differing arguments",
		},
		{
			name: "mutually exclusive types through a shared fragment",
			query: `{ result { ...Shared ...Shared } }
				fragment Shared on Result {
					... on A { same: a }
					... on B { same: b }
				}`,
		},
		{
			name: "shared fragment across operations with different variable defaults",
			query: `query First($arg: Int = 1) { item { ...Shared } }
				query Second($arg: Int = 2) { item { ...Shared } }
				fragment Shared on Item { value(arg: $arg) child { id } }`,
		},
		{
			name: "recursive fragment remains invalid",
			query: `{ item { ...Shared } }
				fragment Shared on Item { id ...Shared }`,
			message: "Cannot spread fragment",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			doc, err := parser.ParseQuery(&ast.Source{Input: tt.query})
			require.NoError(t, err)
			errs := validator.Validate(schema, doc)
			if tt.message == "" {
				require.Empty(t, errs)
			} else {
				require.NotEmpty(t, errs)
				require.Contains(t, errs.Error(), tt.message)
			}
		})
	}
}

func TestFieldCollectionCacheDoesNotSurviveValidation(t *testing.T) {
	schema := gqlparser.MustLoadSchema(&ast.Source{Input: fieldCollectionSchema})
	doc, err := parser.ParseQuery(&ast.Source{Input: `
		{ item { ...Shared name } }
		fragment Shared on Item { id }
	`})
	require.NoError(t, err)
	require.Empty(t, validator.Validate(schema, doc))

	// Keep the selection slice identity and length while changing its response
	// name between passes. A cache surviving Validate would miss this conflict.
	field := doc.Fragments.ForName("Shared").SelectionSet[0].(*ast.Field)
	field.Alias = "name"
	errs := validator.Validate(schema, doc)
	require.NotEmpty(t, errs)
	require.Contains(t, errs.Error(), `"name" and "id" are different fields`)
}

func BenchmarkParseAndValidateSharedFragments(b *testing.B) {
	schema := gqlparser.MustLoadSchema(&ast.Source{Input: fieldCollectionSchema})
	for _, count := range []int{1, 16, 64} {
		b.Run(fmt.Sprintf("paths_%d", count), func(b *testing.B) {
			var query strings.Builder
			query.WriteString("{")
			for i := 0; i < count; i++ {
				fmt.Fprintf(&query, " item%d: item { ...Shared ...Nested }", i)
			}
			query.WriteString(`}
				fragment Shared on Item { id name child { ...Leaf } }
				fragment Nested on Item { child { ...Leaf name } }
				fragment Leaf on Item { id name value(arg: 1) }
			`)
			source := &ast.Source{Input: query.String()}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Parsing gives each validation pass a fresh, unbound AST.
				doc, err := parser.ParseQuery(source)
				if err != nil {
					b.Fatal(err)
				}
				if errs := validator.Validate(schema, doc); len(errs) != 0 {
					b.Fatal(errs)
				}
			}
		})
	}
}
