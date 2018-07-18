package ast

type DefinitionKind string

const (
	Scalar      DefinitionKind = "SCALAR"
	Object      DefinitionKind = "OBJECT"
	Interface   DefinitionKind = "INTERFACE"
	Union       DefinitionKind = "UNION"
	Enum        DefinitionKind = "ENUM"
	InputObject DefinitionKind = "INPUT_OBJECT"
)

// ObjectDefinition is the core type definition object, it includes all of the definable types
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
	Directives  Directives
	Interfaces  []string              // object and input object
	Fields      FieldList             // object and input object
	Types       []string              // union
	Values      []EnumValueDefinition // enum
}

func (d *Definition) Field(name string) *FieldDefinition {
	for _, f := range d.Fields {
		if f.Name == name {
			return f
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
	DefaultValue *Value    // only for input objects
	Type         *Type
	Directives   Directives
}

type FieldList []*FieldDefinition

func (f FieldList) ForName(name string) *FieldDefinition {
	for _, field := range f {
		if field.Name == name {
			return field
		}
	}
	return nil
}

type EnumValueDefinition struct {
	Description string
	Name        string
	Directives  Directives
}

// Directive Definitions

type DirectiveDefinition struct {
	Description string
	Name        string
	Arguments   FieldList
	Locations   []DirectiveLocation
}
