package ast

type Source struct {
	Name  string
	Input string
}

type Location struct {
	Line   int
	Column int
}
