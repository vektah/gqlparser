package ast

import "strconv"

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

type Variable string
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
