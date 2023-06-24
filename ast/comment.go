package ast

import "strings"

type Comment struct {
	Text     string
	Position *Position
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
	if len(c.List) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, comment := range c.List {
		builder.WriteString(comment.Text)
	}
	return builder.String()
}
