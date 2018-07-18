package ast

type QueryDocument struct {
	Operations OperationList
	Fragments  FragmentDefinitionList
}

type SchemaDocument struct {
	Schema          SchemaDefinitionList
	SchemaExtension SchemaDefinitionList
	Directives      DirectiveDefinitionList
	Definitions     DefinitionList
	Extensions      DefinitionList
}

type SchemaDefinition struct {
	Description    string
	Directives     DirectiveList
	OperationTypes OperationTypeDefinitionList
}

type OperationTypeDefinition struct {
	Operation Operation
	Type      string
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
			defs = append(defs, s.Types[t])
		}
		return defs
	}

	return s.PossibleTypes[def.Name]
}
