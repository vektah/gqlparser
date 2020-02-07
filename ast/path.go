package ast

import (
	"bytes"
	"fmt"
)

type Path []PathElement

type PathElement interface {
	isPathElement()
}

var _ PathElement = PathIndex(0)
var _ PathElement = PathName("")

func (path Path) String() string {
	var str bytes.Buffer
	for i, v := range path {

		switch v := v.(type) {
		case PathIndex:
			str.WriteString(fmt.Sprintf("[%d]", v))
		case PathName:
			if i != 0 {
				str.WriteByte('.')
			}
			str.WriteString(string(v))
		default:
			panic(fmt.Sprintf("unknown type: %T", v))
		}
	}
	return str.String()
}

type PathIndex int

func (_ PathIndex) isPathElement() {}

type PathName string

func (_ PathName) isPathElement() {}
