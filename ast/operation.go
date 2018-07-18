package ast

type Operation string

const (
	Query        Operation = "query"
	Mutation     Operation = "mutation"
	Subscription Operation = "subscription"
)

type OperationDefinition struct {
	Operation           Operation
	Name                string
	VariableDefinitions VariableDefinitionList
	Directives          DirectiveList
	SelectionSet        SelectionSet
}

type VariableDefinition struct {
	Variable     string
	Type         *Type
	DefaultValue *Value

	// Requires validation
	Definition *Definition
	Used       bool `dump:"-"`
}
