package ast

import (
	"encoding/json"
	"fmt"
	"strconv"
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

func (v *Value) Value(vars map[string]interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	switch v.Kind {
	case Variable:
		if value, ok := vars[v.Raw]; ok {
			return value, nil
		}
		if v.VariableDefinition != nil && v.VariableDefinition.DefaultValue != nil {
			return v.VariableDefinition.DefaultValue.Value(vars)
		}
		return nil, nil
	case IntValue:
		return strconv.ParseInt(v.Raw, 10, 64)
	case FloatValue:
		return strconv.ParseFloat(v.Raw, 64)
	case StringValue, BlockValue, EnumValue:
		return v.Raw, nil
	case BooleanValue:
		return strconv.ParseBool(v.Raw)
	case NullValue:
		return nil, nil
	case ListValue:
		var val []interface{}
		for _, elem := range v.Children {
			elemVal, err := elem.Value.Value(vars)
			if err != nil {
				return val, err
			}
			val = append(val, elemVal)
		}
		return val, nil
	case ObjectValue:
		val := map[string]interface{}{}
		for _, elem := range v.Children {
			elemVal, err := elem.Value.Value(vars)
			if err != nil {
				return val, err
			}
			val[elem.Name] = elemVal
		}
		return val, nil
	default:
		panic(fmt.Errorf("unknown value kind %d", v.Kind))
	}
}

func (v *Value) String() string {
	switch v.Kind {
	case Variable, IntValue, FloatValue, EnumValue, BooleanValue, NullValue:
		return v.Raw
	case StringValue, BlockValue:
		return strconv.Quote(v.Raw)
	case ListValue:
		return "list"
	case ObjectValue:
		return "object"
	default:
		panic(fmt.Errorf("unknown value kind %d", v.Kind))
	}
}

func (v *Value) Dump() string {
	if v == nil {
		return "<nil>"
	}
	if v.Kind == Variable {
		return "$" + v.Raw
	}
	val, _ := v.Value(nil)
	enc, _ := json.Marshal(val)
	return string(enc)
}

func (v ValueKind) String() string {
	switch v {
	case Variable:
		return "Variable"
	case IntValue:
		return "IntValue"
	case FloatValue:
		return "FloatValue"
	case StringValue:
		return "StringValue"
	case BlockValue:
		return "BlockValue"
	case BooleanValue:
		return "BooleanValue"
	case NullValue:
		return "NullValue"
	case EnumValue:
		return "EnumValue"
	case ListValue:
		return "ListValue"
	case ObjectValue:
		return "ObjectValue"
	}
	return "Unknown"
}
