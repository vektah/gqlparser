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
	VariableDefinitions VariableDefinitions
	Directives          []*Directive
	SelectionSet        SelectionSet
}

type VariableDefinitions []VariableDefinition

func (v VariableDefinitions) Find(name string) *VariableDefinition {
	for i := range v {
		def := v[i]
		if string(def.Variable) == name {
			return &def
		}
	}
	return nil
}

type VariableDefinition struct {
	Variable     Variable
	Type         Type
	DefaultValue Value
}
