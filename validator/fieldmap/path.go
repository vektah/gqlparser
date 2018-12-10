package fieldmap

import (
	"strings"

	"github.com/vektah/gqlparser/ast"
)

type Path []interface{}

func (p Path) Equal(other Path) bool {
	ai := 0
	bi := 0
	for {
		if ai >= len(p) && bi >= len(other) {
			return true
		}
		if ai >= len(p) && bi >= len(other) {
			return false
		}

		a := p[ai]
		b := other[bi]

		fieldA, aIsField := a.(*ast.Field)
		fieldB, bIsField := b.(*ast.Field)

		if aIsField && bIsField && fieldA.Alias == fieldB.Alias {
			ai++
			bi++
			continue
		}

		fragA, aIsFrag := a.(*ast.InlineFragment)
		fragB, bIsFrag := b.(*ast.InlineFragment)
		if aIsFrag && fragA.TypeCondition == "" {
			ai++
			continue
		}

		if bIsFrag && fragB.TypeCondition == "" {
			bi++
			continue
		}

		if aIsFrag && bIsFrag && fragA.TypeCondition == fragB.TypeCondition {
			ai++
			bi++
			continue
		}

		return false
	}
}

func (p Path) String() string {
	var parts []string
	for _, v := range p {
		switch v := v.(type) {
		case *ast.Field:
			parts = append(parts, v.Alias)
		case *ast.InlineFragment:
			parts = append(parts, "on "+v.TypeCondition)
		}
	}
	return strings.Join(parts, ".")
}
