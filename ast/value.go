package ast

import (
	"fmt"
	"strconv"
	"strings"
)

type ValueKind int

const (
	Variable ValueKind = iota
	IntValue
	FloatValue
	StringValue
	BlockValue
	BooleanValue
	NullValue
	EnumValue
	ListValue
	ObjectValue
)

type Value struct {
	Raw      string
	Children ChildValueList
	Kind     ValueKind
	Position *Position `dump:"-"`

	// Require validation
	Definition         *Definition
	VariableDefinition *VariableDefinition
	ExpectedType       *Type
}

type ChildValue struct {
	Name     string
	Value    *Value
	Position *Position `dump:"-"`
}

func (v *Value) String() string {
	if v == nil {
		return "<nil>"
	}
	switch v.Kind {
	case Variable:
		return "$" + v.Raw
	case IntValue, FloatValue, EnumValue, BooleanValue, NullValue:
		return v.Raw
	case StringValue, BlockValue:
		return strconv.Quote(v.Raw)
	case ListValue:
		var val []string
		for _, elem := range v.Children {
			val = append(val, elem.Value.String())
		}
		return "[" + strings.Join(val, ",") + "]"
	case ObjectValue:
		var val []string
		for _, elem := range v.Children {
			val = append(val, strconv.Quote(elem.Name)+":"+elem.Value.String())
		}
		return "{" + strings.Join(val, ",") + "}"
	default:
		panic(fmt.Errorf("unknown value kind %d", v.Kind))
	}
}

func (v *Value) Dump() string {
	return v.String()
}
