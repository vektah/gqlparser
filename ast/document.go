package ast

type QueryDocument struct {
	Operations []OperationDefinition
	Fragments  []FragmentDefinition
}

func (d QueryDocument) GetOperation(name string) *OperationDefinition {
	for _, o := range d.Operations {
		if o.Name == name {
			return &o
		}
	}
	return nil
}

func (d QueryDocument) GetFragment(name string) *FragmentDefinition {
	for _, f := range d.Fragments {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

type SchemaDocument struct {
	Schema          []SchemaDefinition
	SchemaExtension []SchemaDefinition
	Directives      []DirectiveDefinition
	Definitions     []Definition
	Extensions      []Definition
}

type SchemaDefinition struct {
	Description    string
	Directives     []*Directive
	OperationTypes []OperationTypeDefinition
}

type OperationTypeDefinition struct {
	Operation Operation
	Type      NamedType
}

type Schema struct {
	Query        *Definition
	Mutation     *Definition
	Subscription *Definition

	Types      map[string]*Definition
	Directives map[string]*DirectiveDefinition

	PossibleTypes map[string][]*Definition
}

func (s *Schema) AddPossibleType(name string, def *Definition) {
	s.PossibleTypes[name] = append(s.PossibleTypes[name], def)
}

// GetPossibleTypes will enumerate all the definitions for a given interface or union
func (s *Schema) GetPossibleTypes(def *Definition) []*Definition {
	if def.Kind == Union {
		var defs []*Definition
		for _, t := range def.Types {
			defs = append(defs, s.Types[t.Name()])
		}
		return defs
	}

	return s.PossibleTypes[def.Name]
}
