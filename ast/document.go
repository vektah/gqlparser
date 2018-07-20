package ast

type QueryDocument struct {
	Operations OperationList
	Fragments  FragmentDefinitionList
	Position   *Position `dump:"-"`
}

type SchemaDocument struct {
	Schema          SchemaDefinitionList
	SchemaExtension SchemaDefinitionList
	Directives      DirectiveDefinitionList
	Definitions     DefinitionList
	Extensions      DefinitionList
	Position        *Position `dump:"-"`
}

func (d *SchemaDocument) Merge(other *SchemaDocument) {
	d.Schema = append(d.Schema, other.Schema...)
	d.SchemaExtension = append(d.SchemaExtension, other.SchemaExtension...)
	d.Directives = append(d.Directives, other.Directives...)
	d.Definitions = append(d.Definitions, other.Definitions...)
	d.Extensions = append(d.Extensions, other.Extensions...)
}

type SchemaDefinition struct {
	Description    string
	Directives     DirectiveList
	OperationTypes OperationTypeDefinitionList
	Position       *Position `dump:"-"`
}

type OperationTypeDefinition struct {
	Operation Operation
	Type      string
	Position  *Position `dump:"-"`
}
