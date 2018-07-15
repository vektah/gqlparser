package ast

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
