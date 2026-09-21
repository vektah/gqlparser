package rules

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2/ast"
)

func TestFieldCollectionCacheSelectionIdentity(t *testing.T) {
	selections := ast.SelectionSet{&ast.Field{Name: "a"}, &ast.Field{Name: "b"}}
	tests := []struct {
		name   string
		first  ast.SelectionSet
		second ast.SelectionSet
		reuse  bool
	}{
		{"same selection", selections, selections, true},
		{"shared backing with different length", selections, selections[:1], false},
		{"different starting slot", selections[:1], selections[1:], false},
		{"different backing", selections, append(ast.SelectionSet(nil), selections...), false},
		{"empty selections", nil, ast.SelectionSet{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &overlappingFieldsCanBeMergedManager{
				cachedFieldsAndFragmentNames: make(map[selectionSetKey]fieldCollection),
			}
			first, _ := m.getFieldsAndFragmentNames(tt.first)
			second, _ := m.getFieldsAndFragmentNames(tt.second)
			if tt.reuse {
				require.Same(t, first, second)
			} else {
				require.NotSame(t, first, second)
			}
			expected, fragments := collectFieldsAndFragmentNames(tt.second)
			require.Equal(t, expected, second)
			_, cachedFragments := m.getFieldsAndFragmentNames(tt.second)
			require.Equal(t, fragments, cachedFragments)
		})
	}
}

func TestFieldCollectionCachePreservesOrderAndReferences(t *testing.T) {
	first := &ast.Field{Name: "name", Alias: "label"}
	second := &ast.Field{Name: "id"}
	third := &ast.Field{Name: "otherName", Alias: "label"}
	spread := &ast.FragmentSpread{Name: "Shared"}
	selections := ast.SelectionSet{
		first,
		&ast.InlineFragment{SelectionSet: ast.SelectionSet{second, third, spread}},
	}
	m := &overlappingFieldsCanBeMergedManager{
		cachedFieldsAndFragmentNames: make(map[selectionSetKey]fieldCollection),
	}
	fields, fragments := m.getFieldsAndFragmentNames(selections)
	require.Equal(t, []string{"label", "id"}, fields.seq)
	require.Equal(t, []*ast.Field{first, third}, fields.data["label"])
	require.Equal(t, []*ast.FragmentSpread{spread}, fragments)

	// The walker binds definitions on the AST; cached collections must retain
	// those nodes rather than copying their partially populated contents.
	definition := &ast.FieldDefinition{Name: "name"}
	first.Definition = definition
	fragmentDefinition := &ast.FragmentDefinition{Name: "Shared"}
	spread.Definition = fragmentDefinition
	cachedFields, cachedFragments := m.getFieldsAndFragmentNames(selections)
	require.Same(t, fields, cachedFields)
	require.Same(t, definition, cachedFields.data["label"][0].Definition)
	require.Same(t, fragmentDefinition, cachedFragments[0].Definition)
}
