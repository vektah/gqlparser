package ast

func NonNullNamedType(named string) *Type {
	return &Type{NamedType: named, NonNull: true}
}

func NamedType(named string) *Type {
	return &Type{NamedType: named, NonNull: false}
}

func NonNullListType(elem *Type) *Type {
	return &Type{Elem: elem, NonNull: true}
}

func ListType(elem *Type) *Type {
	return &Type{Elem: elem, NonNull: false}
}

type Type struct {
	NamedType string
	Elem      *Type
	NonNull   bool
}

func (t *Type) Name() string {
	if t.NamedType != "" {
		return t.NamedType
	}

	return t.Elem.Name()
}

func (t *Type) String() string {
	nn := ""
	if t.NonNull {
		nn = "!"
	}
	if t.NamedType != "" {
		return t.NamedType + nn
	}

	return "[" + t.Elem.String() + "]" + nn
}

func (t *Type) IsCompatible(other *Type) bool {
	if t.NamedType != other.NamedType {
		return false
	}

	if t.Elem != nil && other.Elem == nil {
		return false
	}

	if t.Elem != nil && !t.Elem.IsCompatible(other.Elem) {
		return false
	}

	if other.NonNull {
		return t.NonNull
	}

	return true
}

func (v *Type) Dump() string {
	return v.String()
}
