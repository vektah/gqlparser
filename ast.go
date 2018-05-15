package gqlparser

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
	Schema          []SchemaDefinition
	SchemaExtension []SchemaDefinition
	Directives      []DirectiveDefinition
	Definitions     []Definition
	Extensions      []Definition
}

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

func (n NamedType) Name() string {
	return string(n)
}

type ListType struct {
	Type Type
}

type NonNullType struct {
	Type Type
}

type SchemaDefinition struct {
	Description    string
	Directives     []Directive
	OperationTypes []OperationTypeDefinition
}

type OperationTypeDefinition struct {
	Operation Operation
	Type      NamedType
}

type DefinitionKind string

const (
	Scalar      DefinitionKind = "SCALAR"
	Object      DefinitionKind = "OBJECT"
	Interface   DefinitionKind = "INTERFACE"
	Union       DefinitionKind = "UNION"
	Enum        DefinitionKind = "ENUM"
	InputObject DefinitionKind = "INPUT_OBJECT"
)

// Definition is the core type definition object, it includes all of the definable types
// but does *not* cover schema or directives.
//
// @vektah: Javascript implementation has different types for all of these, but they are
// more similar than different and don't define any behaviour. I think this style of
// "some hot" struct works better, at least for go.
//
// Type extensions are also represented by this same struct.
type Definition struct {
	Kind        DefinitionKind
	Description string
	Name        string
	Directives  []Directive
	Interfaces  []NamedType           // object and input object
	Fields      []FieldDefinition     // object and input object
	Types       []NamedType           // union
	Values      []EnumValueDefinition // enum
}

func (d *Definition) Field(name string) *FieldDefinition {
	for _, f := range d.Fields {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

type FieldDefinition struct {
	Description  string
	Name         string
	Arguments    []FieldDefinition // only for objects
	DefaultValue Value             // only for input objects
	Type         Type
	Directives   []Directive
}

type EnumValueDefinition struct {
	Description string
	Name        string
	Directives  []Directive
}

// Directive Definitions

type DirectiveDefinition struct {
	Description string
	Name        string
	Arguments   []FieldDefinition
	Locations   []DirectiveLocation
}
