package formatter

import "github.com/vektah/gqlparser/v2/ast"

func getComment(node interface{}) *ast.CommentGroup {
	switch n := node.(type) {
	case *ast.Field:
		return n.Comment
	case *ast.Argument:
		return n.Comment
	case *ast.QueryDocument:
		return n.Comment
	case *ast.SchemaDocument:
		return n.Comment
	case *ast.FragmentSpread:
		return n.Comment
	case *ast.InlineFragment:
		return n.Comment
	case *ast.OperationDefinition:
		return n.Comment
	case *ast.VariableDefinition:
		return n.Comment
	case *ast.Value:
		return n.Comment
	case *ast.ChildValue:
		return n.Comment
	case *ast.FieldDefinition:
		return n.Comment
	case *ast.ArgumentDefinition:
		return n.Comment
	case *ast.EnumValueDefinition:
		return n.Comment
	case *ast.DirectiveDefinition:
		return n.Comment
	}
	return nil
}
