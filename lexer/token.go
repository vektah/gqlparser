package lexer

import (
	"fmt"
	"strconv"
)

const (
	Invalid Type = iota
	EOF
	Bang
	Dollar
	Amp
	ParenL
	ParenR
	Spread
	Colon
	Equals
	At
	BracketL
	BracketR
	BraceL
	BraceR
	Pipe
	Name
	Int
	Float
	String
	BlockString
	Comment
)

func (t Type) String() string {
	switch t {
	case Invalid:
		return "Invalid"
	case EOF:
		return "EOF"
	case Bang:
		return "Bang"
	case Dollar:
		return "Dollar"
	case Amp:
		return "Amp"
	case ParenL:
		return "ParenL"
	case ParenR:
		return "ParenR"
	case Spread:
		return "Spread"
	case Colon:
		return "Colon"
	case Equals:
		return "Equals"
	case At:
		return "At"
	case BracketL:
		return "BracketL"
	case BracketR:
		return "BracketR"
	case BraceL:
		return "BraceL"
	case BraceR:
		return "BraceR"
	case Pipe:
		return "Pipe"
	case Name:
		return "Name"
	case Int:
		return "Int"
	case Float:
		return "Float"
	case String:
		return "String"
	case BlockString:
		return "BlockString"
	case Comment:
		return "Comment"
	}
	return "Unknown " + strconv.Itoa(int(t))
}

// Kind represents a type of token. The types are predefined as constants.
type Type int

type Token struct {
	Kind   Type   // The token type.
	Value  string // The literal value consumed.
	Start  int    // The starting position, in runes, of this token in the input.
	End    int    // The end position, in runes, of this token in the input.
	Line   int    // The line number at the start of this item.
	Column int    // The line number at the start of this item.
}

func (t Token) String() string {
	return fmt.Sprintf("%s[%s, line: %d, column: %d]", t.Kind, strconv.Quote(t.Value), t.Line, t.Column)
}
