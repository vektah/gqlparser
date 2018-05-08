package parser

type Operation string

const (
	Query        Operation = "query"
	Mutation     Operation = "mutation"
	Subscription Operation = "subscription"
)

type DirectiveLocation string

const (
	// Executable
	LocationQuery              DirectiveLocation = `QUERY`
	LocationMutation           DirectiveLocation = `MUTATION`
	LocationSubscription       DirectiveLocation = `SUBSCRIPTION`
	LocationField              DirectiveLocation = `FIELD`
	LocationFragmentDefinition DirectiveLocation = `FRAGMENT_DEFINITION`
	LocationFragmentSpread     DirectiveLocation = `FRAGMENT_SPREAD`
	LocationInlineFragment     DirectiveLocation = `INLINE_FRAGMENT`

	// Type System
	LocationSchema               DirectiveLocation = `SCHEMA`
	LocationScalar               DirectiveLocation = `SCALAR`
	LocationObject               DirectiveLocation = `OBJECT`
	LocationFieldDefinition      DirectiveLocation = `FIELD_DEFINITION`
	LocationArgumentDefinition   DirectiveLocation = `ARGUMENT_DEFINITION`
	LocationIinterface           DirectiveLocation = `INTERFACE`
	LocationUnion                DirectiveLocation = `UNION`
	LocationEnum                 DirectiveLocation = `ENUM`
	LocationEnumValue            DirectiveLocation = `ENUM_VALUE`
	LocationInputObject          DirectiveLocation = `INPUT_OBJECT`
	LocationInputFieldDefinition DirectiveLocation = `INPUT_FIELD_DEFINITION`
)

// Document

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
	Definitions []Definition
	Extensions  []TypeExtension
}

type Definition interface {
	isDefinition()
}

func (OperationDefinition) isDefinition()       {}
func (FragmentDefinition) isDefinition()        {}
func (SchemaExtension) isDefinition()           {}
func (SchemaDefinition) isDefinition()          {}
func (DirectiveDefinition) isDefinition()       {}
func (ScalarTypeDefinition) isDefinition()      {}
func (ObjectTypeDefinition) isDefinition()      {}
func (InterfaceTypeDefinition) isDefinition()   {}
func (UnionTypeDefinition) isDefinition()       {}
func (EnumTypeDefinition) isDefinition()        {}
func (InputObjectTypeDefinition) isDefinition() {}
func (ScalarTypeExtension) isDefinition()       {}
func (ObjectTypeExtension) isDefinition()       {}
func (InterfaceTypeExtension) isDefinition()    {}
func (UnionTypeExtension) isDefinition()        {}
func (EnumTypeExtension) isDefinition()         {}
func (InputObjectTypeExtension) isDefinition()  {}

type OperationDefinition struct {
	Operation           Operation
	Name                string
	VariableDefinitions []VariableDefinition
	Directives          []Directive
	SelectionSet        SelectionSet
}

type VariableDefinition struct {
	Variable     Variable
	Type         Type
	DefaultValue Value
}

type Variable string

type SelectionSet []Selection

type Selection interface {
	isSelection()
}

func (Field) isSelection()          {}
func (FragmentSpread) isSelection() {}
func (InlineFragment) isSelection() {}

type Field struct {
	Alias        string
	Name         string
	Arguments    []Argument
	Directives   []Directive
	SelectionSet SelectionSet
}

type Argument struct {
	Name  string
	Value Value
}

// Fragments

type FragmentSpread struct {
	Name       string
	Directives []Directive
}

type InlineFragment struct {
	TypeCondition NamedType
	Directives    []Directive
	SelectionSet  SelectionSet
}

type FragmentDefinition struct {
	Name string
	// Note: fragment variable definitions are experimental and may be changed
	// or removed in the future.
	VariableDefinition []VariableDefinition
	TypeCondition      NamedType
	Directives         []Directive
	SelectionSet       SelectionSet
}

// Values

type Value interface {
	isValue()
}

func (Variable) isValue()     {}
func (IntValue) isValue()     {}
func (FloatValue) isValue()   {}
func (StringValue) isValue()  {}
func (BlockValue) isValue()   {}
func (BooleanValue) isValue() {}
func (NullValue) isValue()    {}
func (EnumValue) isValue()    {}
func (ListValue) isValue()    {}
func (ObjectValue) isValue()  {}

type IntValue string
type FloatValue string
type StringValue string
type BlockValue string
type BooleanValue bool
type NullValue struct{}
type EnumValue string
type ListValue []Value
type ObjectValue []ObjectField

type ObjectField struct {
	Name  string
	Value Value
}

// Directives

type Directive struct {
	Name      string
	Arguments []Argument
}

// Type Reference

type Type interface {
	isType()
}

func (NamedType) isType()   {}
func (ListType) isType()    {}
func (NonNullType) isType() {}

type NamedType string

type ListType struct {
	Type Type
}

type NonNullType struct {
	Type Type
}

// Type System Definition

type TypeDefinition interface {
	isTypeDefinition()
}

func (SchemaDefinition) isTypeDefinition()          {}
func (DirectiveDefinition) isTypeDefinition()       {}
func (ScalarTypeDefinition) isTypeDefinition()      {}
func (ObjectTypeDefinition) isTypeDefinition()      {}
func (InterfaceTypeDefinition) isTypeDefinition()   {}
func (UnionTypeDefinition) isTypeDefinition()       {}
func (EnumTypeDefinition) isTypeDefinition()        {}
func (InputObjectTypeDefinition) isTypeDefinition() {}

type SchemaDefinition struct {
	Description    string
	Directives     []Directive
	OperationTypes []OperationTypeDefinition
}

type OperationTypeDefinition struct {
	Operation Operation
	Type      NamedType
}

// Type Definition

type ScalarTypeDefinition struct {
	Description string
	Name        string
	Directives  []Directive
}

type ObjectTypeDefinition struct {
	Description string
	Name        string
	Interfaces  []NamedType
	Directives  []Directive
	Fields      []FieldDefinition
}

type FieldDefinition struct {
	Description string
	Name        string
	Arguments   []InputValueDefinition
	Type        Type
	Directives  []Directive
}

type InputValueDefinition struct {
	Description  string
	Name         string
	Type         Type
	DefaultValue Value
	Directives   []Directive
}

type InterfaceTypeDefinition struct {
	Description string
	Name        string
	Directives  []Directive
	Fields      []FieldDefinition
}

type UnionTypeDefinition struct {
	Description string
	Name        string
	Directives  []Directive
	Types       []NamedType
}

type EnumTypeDefinition struct {
	Description string
	Name        string
	Directives  []Directive
	Values      []EnumValueDefinition
}

type EnumValueDefinition struct {
	Description string
	Name        string
	Directives  []Directive
}

type InputObjectTypeDefinition struct {
	Description string
	Name        string
	Directives  []Directive
	Fields      []InputValueDefinition
}

// Directive Definitions

type DirectiveDefinition struct {
	Description string
	Name        string
	Arguments   []InputValueDefinition
	Locations   []DirectiveLocation
}

type TypeExtension interface {
	isTypeExtension()
}

func (SchemaExtension) isTypeExtension()          {}
func (ScalarTypeExtension) isTypeExtension()      {}
func (ObjectTypeExtension) isTypeExtension()      {}
func (InterfaceTypeExtension) isTypeExtension()   {}
func (UnionTypeExtension) isTypeExtension()       {}
func (EnumTypeExtension) isTypeExtension()        {}
func (InputObjectTypeExtension) isTypeExtension() {}

// Type System Extensions

type SchemaExtension struct {
	Description    string
	Directives     []Directive
	OperationTypes []OperationTypeDefinition
}

// Type Extensions

type ScalarTypeExtension struct {
	Description string
	Name        string
	Directives  []Directive
}

type ObjectTypeExtension struct {
	Description string
	Name        string
	Interfaces  []NamedType
	Directives  []Directive
	Fields      []FieldDefinition
}

type InterfaceTypeExtension struct {
	Description string
	Name        string
	Directives  []Directive
	Fields      []FieldDefinition
}

type UnionTypeExtension struct {
	Description string
	Name        string
	Directives  []Directive
	Types       []NamedType
}

type EnumTypeExtension struct {
	Description string
	Name        string
	Directives  []Directive
	Values      []EnumValueDefinition
}

type InputObjectTypeExtension struct {
	Description string
	Name        string
	Directives  []Directive
	Fields      []InputValueDefinition
}
