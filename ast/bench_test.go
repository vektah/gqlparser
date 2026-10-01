package ast_test

import (
	"strconv"
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
)

// BenchmarkFieldListForName measures the lookup the validator performs for every field of
// every selection set: a linear scan of the parent type's fields.
//
// Linear is the right shape for the handful of fields most types have, and the wrong shape
// for the ones that have hundreds — an introspection type, or a generated schema's root
// Query. The sizes below bracket that, and the position of the name within the list decides
// how much of the scan runs, so a hit at the end and a miss are measured apart from a hit at
// the front. A change that replaced the scan with a map would show here first.
func BenchmarkFieldListForName(b *testing.B) {
	for _, size := range []int{8, 64, 512} {
		fields := make(ast.FieldList, size)
		for i := range fields {
			fields[i] = &ast.FieldDefinition{Name: "field" + strconv.Itoa(i)}
		}

		lookups := map[string]string{
			"first":   "field0",
			"last":    "field" + strconv.Itoa(size-1),
			"missing": "absent",
		}
		for _, position := range []string{"first", "last", "missing"} {
			name := lookups[position]
			b.Run(strconv.Itoa(size)+"/"+position, func(b *testing.B) {
				b.ReportAllocs()

				for b.Loop() {
					fields.ForName(name)
				}
			})
		}
	}
}

// BenchmarkDirectiveListForName measures the other scan on the hot path: every field,
// fragment and operation is checked for directives, and most carry none. The empty case is
// the one that runs most often.
func BenchmarkDirectiveListForName(b *testing.B) {
	empty := ast.DirectiveList{}
	populated := ast.DirectiveList{
		{Name: "include"}, {Name: "skip"}, {Name: "deprecated"},
	}

	b.Run("empty", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			empty.ForName("skip")
		}
	})
	b.Run("populated", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			populated.ForName("skip")
		}
	})
}

// BenchmarkTypeString measures rendering a type reference, which every type-mismatch message
// does twice and which the validator calls while comparing field types.
func BenchmarkTypeString(b *testing.B) {
	named := &ast.Type{NamedType: "String"}
	nonNullList := &ast.Type{
		Elem:    &ast.Type{Elem: &ast.Type{NamedType: "Episode", NonNull: true}, NonNull: true},
		NonNull: true,
	}

	b.Run("named", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = named.String()
		}
	})
	b.Run("nested non-null list", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = nonNullList.String()
		}
	})
}
