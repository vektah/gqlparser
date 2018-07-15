package ast

type FragmentSpread struct {
	Name       string
	Directives []Directive

	// Require validation
	ObjectDefinition *Definition
	Definition       *FragmentDefinition
}

type InlineFragment struct {
	TypeCondition NamedType
	Directives    []Directive
	SelectionSet  SelectionSet

	// Require validation
	ObjectDefinition *Definition
}

type FragmentDefinition struct {
	Name string
	// Note: fragment variable definitions are experimental and may be changed
	// or removed in the future.
	VariableDefinition []VariableDefinition
	TypeCondition      NamedType
	Directives         []Directive
	SelectionSet       SelectionSet

	// Require validation
	Definition *Definition
}
