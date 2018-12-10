package validator

import (
	"bytes"
	"strings"

	"github.com/vektah/gqlparser/ast"
	. "github.com/vektah/gqlparser/validator"
	"github.com/vektah/gqlparser/validator/fieldmap"
)

func init() {

	AddRule("OverlappingFieldsCanBeMerged", func(observers *Events, addError AddErrFunc) {
		var result fieldmap.Map
		var keyPath []string
		var typePath []interface{}

		observers.OnEnterField(func(walker *Walker, field *ast.Field) {
			keyPath = append(keyPath, field.Alias)
			typePath = append(typePath, field)
		})

		observers.OnExitField(func(walker *Walker, field *ast.Field) {
			result.Push(keyStr(keyPath), &fieldmap.Node{Field: field, Path: typePath})
			keyPath = keyPath[0 : len(keyPath)-1]
			typePath = typePath[0 : len(typePath)-1]
		})

		checkOverlappingFields := func() {
			//fmt.Println()
			//fmt.Println("All fields:")
			//fmt.Printf("%-30s %-10s %s \n", "KEY", "ALIAS", "NAME")
			//for pos := 0; pos < result.Len(); pos++ {
			//	key, overlappingFields := result.At(pos)
			//	for _, f := range overlappingFields {
			//		fmt.Printf("%-30s %-10s %s \n", key, f.Alias, f.Name)
			//	}
			//}
			//fmt.Println()

			for pos := 0; pos < result.Len(); pos++ {
				key, overlappingFields := result.At(pos)

				// bitset marking which overlapping fields are in conflict
				// used to dedupe without needing to compare
				fieldConflicts := make([]bool, len(overlappingFields))
				var hasConflicts bool

				for i, outer := range overlappingFields {
					for j := i + 1; j < len(overlappingFields); j++ {
						inner := overlappingFields[j]
						//fmt.Println("CMP", i, j, outer, inner)

						if conflicts(outer, inner) {
							fieldConflicts[i] = true
							fieldConflicts[j] = true
							hasConflicts = true
						}
					}
				}

				if hasConflicts {
					// single pass at the end to grab all the conflicts and build a sensible error message
					var conflictMsg []string
					var pos []*ast.Position
					for i, field := range overlappingFields {
						if !fieldConflicts[i] {
							continue
						}
						conflictMsg = append(conflictMsg, fieldDescription(field.Field))
						pos = append(pos, field.Field.Position)
					}

					addError(
						Message("Field %s has multiple conflicting definitions:\n    %s", key, strings.Join(conflictMsg, "\n    ")),
						At(pos...),
					)
				}
			}
			result.Reset()
		}

		observers.OnEnterInlineFragment(func(walker *Walker, inlineFragment *ast.InlineFragment) {
			typePath = append(typePath, inlineFragment)
		})

		observers.OnExitInlineFragment(func(walker *Walker, inlineFragment *ast.InlineFragment) {
			typePath = typePath[0 : len(typePath)-1]
		})

		observers.OnOperation(func(walker *Walker, operation *ast.OperationDefinition) {
			checkOverlappingFields()
		})

		observers.OnFragment(func(walker *Walker, fragment *ast.FragmentDefinition) {
			checkOverlappingFields()
		})
	})
}

func fieldDescription(field *ast.Field) string {
	var buf bytes.Buffer
	if field.ObjectDefinition != nil {
		buf.WriteString(field.ObjectDefinition.Name)
	} else {
		buf.WriteString("???")
	}
	buf.WriteRune('.')

	buf.WriteString(field.Name)

	buf.WriteRune('(')

	for i, arg := range field.Arguments {
		if i != 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(arg.Name)
		buf.WriteString(": ")
		buf.WriteString(arg.Value.String())
	}

	buf.WriteRune(')')

	for _, dir := range field.Directives {
		buf.WriteRune('@')
		buf.WriteString(dir.Name)

		buf.WriteRune('(')

		for i, arg := range dir.Arguments {
			if i != 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(arg.Name)
			buf.WriteString(": ")
			buf.WriteString(arg.Value.String())
		}

		buf.WriteRune(')')
	}

	buf.WriteRune(' ')
	if field.Definition != nil {
		buf.WriteString(field.Definition.Type.String())
	}

	return buf.String()
}

func conflicts(a, b *fieldmap.Node) bool {
	if a.Field.Definition != nil && b.Field.Definition != nil && !a.Field.Definition.Type.IsCompatible(a.Field.Definition.Type) {
		return true
	}

	if !a.Path.Equal(b.Path) {
		return false
	}

	if a.Field.Name != b.Field.Name {
		return true
	}

	if len(a.Field.Arguments) != len(b.Field.Arguments) {
		return true
	}

	for i := 0; i < len(b.Field.Arguments); i++ {
		if a.Field.Arguments[i].Name != b.Field.Arguments[i].Name {
			return true
		}

		if a.Field.Arguments[i].Value.String() != b.Field.Arguments[i].Value.String() {
			return true
		}
	}

	return false
}

func keyStr(path []string) string {
	var buf bytes.Buffer

	for i, v := range path {
		// skip empty type conditions in inline fragments
		if v == "on " {
			continue
		}

		if i != 0 {
			buf.WriteByte('.')
		}

		buf.WriteString(v)
	}

	return buf.String()
}
