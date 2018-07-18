package ast

type SelectionSet []Selection

type Selection interface {
	isSelection()
}

func (*Field) isSelection()          {}
func (*FragmentSpread) isSelection() {}
func (*InlineFragment) isSelection() {}

type Field struct {
	Alias        string
	Name         string
	Arguments    ArgumentList
	Directives   DirectiveList
	SelectionSet SelectionSet

	// Require validation
	Definition       *FieldDefinition
	ObjectDefinition *Definition
}

type Argument struct {
	Name  string
	Value *Value
}
