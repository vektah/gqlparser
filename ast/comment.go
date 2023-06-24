package ast

import (
	"strconv"
	"strings"
)

type Comment struct {
	Value    string
	Position *Position
}

func (c *Comment) Text() string {
	return strings.TrimSpace(strings.TrimPrefix(c.Value, "#"))
}

type CommentGroup struct {
	List []*Comment
}

func (c *CommentGroup) Start() int {
	if len(c.List) == 0 {
		return 0
	}
	return c.List[0].Position.Start
}

func (c *CommentGroup) End() int {
	if len(c.List) == 0 {
		return 0
	}
	return c.List[len(c.List)-1].Position.End
}

func (c *CommentGroup) Text() string {
	if c == nil || len(c.List) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, comment := range c.List {
		builder.WriteString(comment.Text())
		builder.WriteString("\n")
	}
	return builder.String()
}

func (c *CommentGroup) Dump() string {
	if len(c.List) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, comment := range c.List {
		builder.WriteString(comment.Value)
		builder.WriteString("\n")
	}
	return strconv.Quote(builder.String())
}
