package ast

import "strconv"

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
	LocationInterface            DirectiveLocation = `INTERFACE`
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
	VariableDefinitions VariableDefinitions
	Directives          []Directive
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
	Value(vars map[Variable]interface{}) (interface{}, error)
	String() string
}

func (v Variable) Value(vars map[Variable]interface{}) (interface{}, error) {
	return vars[v], nil
}
func (v IntValue) Value(vars map[Variable]interface{}) (interface{}, error) {
	return strconv.ParseInt(string(v), 10, 64)
}
func (v FloatValue) Value(vars map[Variable]interface{}) (interface{}, error) {
	return strconv.ParseFloat(string(v), 64)
}
func (v StringValue) Value(vars map[Variable]interface{}) (interface{}, error) {
	return string(v), nil
}
func (v BlockValue) Value(vars map[Variable]interface{}) (interface{}, error) {
	return string(v), nil
}
func (v BooleanValue) Value(vars map[Variable]interface{}) (interface{}, error) {
	return bool(v), nil
}
func (v NullValue) Value(vars map[Variable]interface{}) (interface{}, error) {
	return nil, nil
}
func (v EnumValue) Value(vars map[Variable]interface{}) (interface{}, error) {
	return string(v), nil
}
func (v ListValue) Value(vars map[Variable]interface{}) (interface{}, error) {
	var val []interface{}
	for _, elem := range v {
		elemVal, err := elem.Value(vars)
		if err != nil {
			return val, err
		}
		val = append(val, elemVal)
	}
	return val, nil
}
func (v ObjectValue) Value(vars map[Variable]interface{}) (interface{}, error) {
	val := map[string]interface{}{}
	for _, elem := range v {
		elemVal, err := elem.Value.Value(vars)
		if err != nil {
			return val, err
		}
		val[elem.Name] = elemVal
	}
	return val, nil
}

func (v Variable) String() string     { return string(v) }
func (v IntValue) String() string     { return string(v) }
func (v FloatValue) String() string   { return string(v) }
func (v StringValue) String() string  { return strconv.Quote(string(v)) }
func (v BlockValue) String() string   { return strconv.Quote(string(v)) }
func (v BooleanValue) String() string { return strconv.FormatBool(bool(v)) }
func (v NullValue) String() string    { return "null" }
func (v EnumValue) String() string    { return string(v) }
func (v ListValue) String() string    { return "list" }
func (v ObjectValue) String() string  { return "object" }

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

func (o ObjectValue) Find(name string) Value {
	for _, f := range o {
		if f.Name == name {
			return f.Value
		}
	}
	return nil
}

// Directives

type Directive struct {
	Name      string
	Arguments []Argument
}

// Type Reference

type Type interface {
	Name() string
	String() string
	IsRequired() bool
	IsCompatible(other Type) bool
}

func (t NamedType) Name() string   { return string(t) }
func (t ListType) Name() string    { return t.Type.Name() }
func (t NonNullType) Name() string { return t.Type.Name() }

func (t NamedType) String() string   { return string(t) }
func (t ListType) String() string    { return "[" + t.Type.String() + "]" }
func (t NonNullType) String() string { return t.Type.String() + "!" }

func (t NamedType) IsRequired() bool   { return false }
func (t ListType) IsRequired() bool    { return false }
func (t NonNullType) IsRequired() bool { return true }

func (t NamedType) IsCompatible(other Type) bool {
	otherType, sameType := other.(NamedType)
	return sameType && otherType == t
}
func (t ListType) IsCompatible(other Type) bool {
	otherType, sameType := other.(ListType)
	return sameType && t.Type.IsCompatible(otherType.Type)
}
func (t NonNullType) IsCompatible(other Type) bool {
	otherType, sameType := other.(NonNullType)
	if sameType {
		return t.Type.IsCompatible(otherType.Type)
	}
	return t.Type.IsCompatible(other)
}

type NamedType string

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
	Fields      FieldList             // object and input object
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

func (d *Definition) EnumValue(name string) *EnumValueDefinition {
	for _, e := range d.Values {
		if e.Name == name {
			return &e
		}
	}
	return nil
}

func (d *Definition) IsLeafType() bool {
	return d.Kind == Enum || d.Kind == Scalar
}

func (d *Definition) IsAbstractType() bool {
	return d.Kind == Interface || d.Kind == Union
}

func (d *Definition) IsCompositeType() bool {
	return d.Kind == Object || d.Kind == Interface || d.Kind == Union
}

func (d *Definition) IsInputType() bool {
	return d.Kind == Scalar || d.Kind == Enum || d.Kind == InputObject
}

func (d *Definition) OneOf(types ...string) bool {
	for _, t := range types {
		if d.Name == t {
			return true
		}
	}
	return false
}

type FieldDefinition struct {
	Description  string
	Name         string
	Arguments    FieldList // only for objects
	DefaultValue Value     // only for input objects
	Type         Type
	Directives   []Directive
}

type FieldList []FieldDefinition

func (f FieldList) ForName(name string) *FieldDefinition {
	for _, field := range f {
		if field.Name == name {
			return &field
		}
	}
	return nil
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
	Arguments   FieldList
	Locations   []DirectiveLocation
}
