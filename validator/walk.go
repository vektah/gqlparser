package validator

import (
	"fmt"

	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/errors"
)

var fieldVisitors []func(vctx *vctx, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field)
var fragmentVisitors []func(vctx *vctx, parentDef *gqlparser.Definition, fragment *gqlparser.FragmentDefinition)
var inlineFragmentVisitors []func(vctx *vctx, parentDef *gqlparser.Definition, inlineFragment *gqlparser.InlineFragment)
var directiveVisitors []func(vctx *vctx, parentDef *gqlparser.Definition, directiveDef *gqlparser.DirectiveDefinition, directive *gqlparser.Directive)
var directiveDecoratedVisitors []func(vctx *vctx, target interface{}, directiveDef *gqlparser.DirectiveDefinition, directive *gqlparser.Directive)

func init() {
	//fieldVisitors = append(fieldVisitors, func(vctx *vctx, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field) {
	//	fmt.Println("ENTER FIELD "+field.Name, parentDef, fieldDef)
	//})
}

type vctx struct {
	schema   *gqlparser.Schema
	document *gqlparser.QueryDocument
	errors   []errors.Validation
}

func (c *vctx) walk() {
	for _, child := range c.document.Operations {
		c.walkOperation(&child)
	}
	for _, child := range c.document.Fragments {
		c.walkFragment(&child)
	}
}

func (c *vctx) walkOperation(operation *gqlparser.OperationDefinition) {
	var def *gqlparser.Definition
	switch operation.Operation {
	case gqlparser.Query, "":
		def = c.schema.Query
	case gqlparser.Mutation:
		def = c.schema.Mutation
	case gqlparser.Subscription:
		def = c.schema.Subscription
	}

	for _, directive := range operation.Directives {
		def := c.schema.Directives[directive.Name]
		for _, v := range directiveDecoratedVisitors {
			v(c, operation, def, &directive)
		}
	}

	for _, v := range operation.SelectionSet {
		c.walkSelection(def, v)
	}
}

func (c *vctx) walkFragment(it *gqlparser.FragmentDefinition) {
	parentDef := c.schema.Types[it.TypeCondition.Name()]
	if parentDef == nil {
		return
	}

	beforeErr := len(c.errors)
	for _, v := range fragmentVisitors {
		v(c, parentDef, it)
	}
	if beforeErr != len(c.errors) {
		return
	}

	for _, child := range it.SelectionSet {
		c.walkSelection(parentDef, child)
	}
}

func (c *vctx) walkDirective(parentDef *gqlparser.Definition, directive *gqlparser.Directive) {
	def := c.schema.Directives[directive.Name]
	for _, v := range directiveVisitors {
		v(c, parentDef, def, directive)
	}
}

func (c *vctx) walkSelection(parentDef *gqlparser.Definition, it gqlparser.Selection) {
	switch it := it.(type) {
	case gqlparser.Field:
		var def *gqlparser.FieldDefinition
		if it.Name == "__typename" {
			def = &gqlparser.FieldDefinition{
				Name: "__typename",
				Type: gqlparser.NamedType("String"),
			}
		} else if parentDef != nil {
			def = parentDef.Field(it.Name)
		}

		for _, v := range fieldVisitors {
			v(c, parentDef, def, &it)
		}

		var nextParentDef *gqlparser.Definition
		if def != nil {
			nextParentDef = c.schema.Types[def.Type.Name()]
		}

		for _, sel := range it.SelectionSet {
			c.walkSelection(nextParentDef, sel)
		}

		for _, directive := range it.Directives {
			def := c.schema.Directives[directive.Name]
			for _, v := range directiveDecoratedVisitors {
				v(c, &it, def, &directive)
			}

			c.walkDirective(nextParentDef, &directive)
		}

	case gqlparser.InlineFragment:
		beforeErr := len(c.errors)
		for _, v := range inlineFragmentVisitors {
			v(c, parentDef, &it)
		}
		if beforeErr != len(c.errors) {
			return
		}

		var nextParentDef *gqlparser.Definition
		if it.TypeCondition.Name() != "" {
			nextParentDef = c.schema.Types[it.TypeCondition.Name()]
		}

		if nextParentDef != nil {
			for _, sel := range it.SelectionSet {
				c.walkSelection(nextParentDef, sel)
			}

			for _, dir := range it.Directives {
				c.walkDirective(nextParentDef, &dir)
			}
		}

	case gqlparser.FragmentSpread:
		for _, directive := range it.Directives {
			def := c.schema.Directives[directive.Name]
			for _, v := range directiveDecoratedVisitors {
				v(c, &it, def, &directive)
			}
		}

	default:
		panic(fmt.Errorf("unsupported %T", it))

	}
}
